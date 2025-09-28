from django.shortcuts import get_object_or_404
from rest_framework import status, generics
from rest_framework.pagination import PageNumberPagination
from rest_framework.permissions import IsAuthenticated
from rest_framework.response import Response
from rest_framework.validators import ValidationError
from expenseapp.models import Account, Transaction
from expenseapp.serializers import TransactionSerializer
from expenseapp.finance import get_curr_dates
from expenseapp.finance import adjust_account_balance, to_date

class TransactionView(generics.ListAPIView): 
    """ Base transaction view to implement pagination """
    permission_classes = [IsAuthenticated]
    serializer_class = TransactionSerializer
    pagination_class = PageNumberPagination


class UserTransactionList(generics.ListCreateAPIView): 
    """ View to handle the list of transactions of a specific user """

    permission_classes = [IsAuthenticated]
    pagination_class = PageNumberPagination

    def get_transaction_response(self, request) -> Response:
        """ Get custom response with pagination of transactions """

        transaction_list = Transaction.objects.filter(user=request.user).order_by("-occur_date")[:50]
        paginator = self.pagination_class()
        transaction_page = paginator.paginate_queryset(transaction_list, request)

        return paginator.get_paginated_response(TransactionSerializer(transaction_page, many=True).data)
    
    def list(self, request, *args, **kwargs) -> Response:
        """ List of 50 latest transactions with pagination """
        return self.get_transaction_response(request)

    def create(self, request, *args, **kwargs) -> Response:
        """ Create new transaction """

        new_trans_serializer = TransactionSerializer(data=request.data)
        if new_trans_serializer.is_valid(): 
            new_transaction = new_trans_serializer.save() # call the create method 

            # adjust balance of the associated account 
            adjust_account_balance(new_transaction.account, new_transaction)

            return self.get_transaction_response(request)
        
        return Response(new_trans_serializer.errors, status=status.HTTP_400_BAD_REQUEST)


class IntervalTransactionList(TransactionView): 
    """
    View to handle the list of transactions between 2 given dates in the endpoints.
    
    Params: first_date, last_date 
    """

    def get_queryset(self):
        # get the params 
        first_date_str = self.request.query_params.get("first_date")
        last_date_str = self.request.query_params.get("last_date")

        # Validate
        if not first_date_str or not last_date_str: 
            raise ValidationError({"message": "First date or last date unspecified"})
        
        # Convert string to actual date obj
        first_date, last_date = to_date(first_date_str), to_date(last_date_str)

        # List of transactions between 2 dates 
        return Transaction.objects.filter(
            user=self.request.user, 
            occur_date__gte=first_date, 
            occur_date__lte=last_date).order_by("-occur_date")
    

class CategoryTransactionList(TransactionView): 
    """
    View to handle the list of latest transactions in each category for a specific user 
    """

    def get_queryset(self):
        # get query param and validate 
        if not self.kwargs['arg_cat']: 
            raise ValidationError({"error": "Category not specified"})
        
        first_date, last_date = get_curr_dates(period_type="month")

        return Transaction.objects.filter(
            user=self.request.user, category=self.kwargs['arg_cat'],
            occur_date__gte=first_date, 
            occur_date__lte=last_date).order_by("-occur_date")
    

class BothTransactionList(TransactionView): 
    """
    View to handle the list of transactions of category between 2 dates.

    Params: first_date, last_date, category
    """ 

    def get_queryset(self):
        # get the query params and validate 
        category = self.request.query_params.get("category")
        if not category:
            raise ValidationError({"message": "Category not specified"})

        first_date_str = self.request.query_params.get("first_date")
        last_date_str = self.request.query_params.get("last_date")
        if not first_date_str or not last_date_str: 
            raise ValidationError({"message": "first date or last date not specified"})
        
        first_date, last_date = to_date(first_date_str), to_date(last_date_str)

        return Transaction.objects.filter(
            user=self.request.user, category=category, 
            occur_date__gte=first_date, 
            occur_date__lte=last_date).order_by("-occur_date")


class AccountTransactionList(TransactionView): 
    """ View to handle the 20 latest transactions of the account """

    def get_queryset(self):
        queried_account = get_object_or_404(Account, pk=self.kwargs["pk"])
        return Transaction.objects.filter(account=queried_account).order_by("-occur_date")[:50]
    

class AccBothTransactionList(TransactionView): 
    """ 
    View to handle the list of transactions in each category for the account between 2 dates

    If dates are not specified, then use 2 dates of latest months
    """

    def get_queryset(self):
        arg_category = self.request.query_params.get("category")
        if not arg_category: 
            raise ValidationError({"error": "Category not specified"})
        
        first_str = self.request.query_params.get("first_date")
        last_str = self.request.query_params.get("last_date")
        if not first_str or not last_str: 
            first_date, last_date = get_curr_dates(period_type="month")
        else: 
            first_date, last_date = to_date(first_str), to_date(last_str)
        
        queried_account = get_object_or_404(Account, pk=self.kwargs['pk'])
        return Transaction.objects.filter(
            account=queried_account, category=arg_category, 
            occur_date__gte=first_date, occur_date__lte=last_date).order_by("-occur_date")
    