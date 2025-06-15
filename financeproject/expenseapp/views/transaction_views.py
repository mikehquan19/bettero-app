from django.shortcuts import get_object_or_404
from rest_framework import status, generics
from rest_framework.pagination import PageNumberPagination
from rest_framework.permissions import IsAuthenticated
from rest_framework.response import Response
from rest_framework.validators import ValidationError
from expenseapp.models import Account, Transaction
from expenseapp.serializers import TransactionSerializer
from expenseapp.finance import get_curr_dates
from datetime import date
from calendar import monthrange
from expenseapp.finance import adjust_account_balance

# base transaction view to handle pagination 
class TransactionView(generics.ListAPIView): 
    permission_classes = [IsAuthenticated]
    serializer_class = TransactionSerializer
    pagination_class = PageNumberPagination


# handling the list of transactions of the user 
class UserTransactionList(generics.ListCreateAPIView): 
    permission_classes = [IsAuthenticated]
    pagination_class = PageNumberPagination

    # get response with pagination of transactions 
    def get_response(self, request):
        transaction_list = Transaction.objects.filter(user=request.user).order_by("-occur_date")[:20]
        paginator = self.pagination_class()
        transaction_page = paginator.paginate_queryset(transaction_list, request)

        return paginator.get_paginated_response(
            TransactionSerializer(transaction_page, many=True).data
        )
    
    # list of 20 latest transactions with pagination
    def list(self, request, *args, **kwargs):
        return self.get_response(request)

    # create new transaction
    def create(self, request, *args, **kwargs):
        new_trans_serializer = TransactionSerializer(data=request.data)
        if new_trans_serializer.is_valid(): 
            new_transaction = new_trans_serializer.save() # call the create method 

            # adjust balance of the associated account 
            adjust_account_balance(new_transaction.account, new_transaction)

            return self.get_response(request)
        
        return Response(new_trans_serializer.errors, status=status.HTTP_400_BAD_REQUEST)


"""
Handling the list of transactions between 2 given dates in the endpoints 
requires pagination 
params: first_date, last_date 
"""
class IntervalTransactionList(TransactionView): 

    def get_queryset(self):
        # get the params 
        first_str = self.request.query_params.get("first_date")
        last_str = self.request.query_params.get("last_date")

        # Validate
        if not first_str or not last_str: 
            raise ValidationError({"message": "First date or last date unspecified"})
        
        # Convert string to actual date obj
        first_list = first_str.split("-")
        last_list = last_str.split("-")

        first_date = date(int(first_list[0]), int(first_list[1]), int(first_list[2]))
        last_date = date(int(last_list[0]), int(last_list[1]), int(last_list[2]))

        # List of transactions between 2 dates 
        return Transaction.objects.filter(
            user=self.request.user, 
            occur_date__gte=first_date, 
            occur_date__lte=last_date).order_by("-occur_date")
    

"""
Handling the list of latest transactions in each category for the user 
requires pagination 
"""
class CategoryTransactionList(TransactionView): 

    def get_queryset(self):
        # get query param and validate 
        if not self.kwargs['arg_cat']: 
            raise ValidationError({"error": "Category not specified"})
        
        first_date, last_date = get_curr_dates(period_type="month")

        return Transaction.objects.filter(
            user=self.request.user, 
            category=self.kwargs['arg_cat'],
            occur_date__gte=first_date, 
            occur_date__lte=last_date).order_by("-occur_date")
    

"""
Handling list of transactions of category between 2 dates
requires pagination 
params: first_date, last_date, category
""" 
class BothTransactionList(TransactionView): 

    def get_queryset(self):
        # get the query params and validate 
        category = self.request.query_params.get("category")
        if not category:
            raise ValidationError({"message": "Category not specified"})

        first_str = self.request.query_params.get("first_date")
        last_str = self.request.query_params.get("last_date")
        if not first_str or not last_str: 
            raise ValidationError({"message": "first date or last date not specified"})
        
        # Process 
        first_list = first_str.split("-")
        last_list = last_str.split("-")

        first_date = date(int(first_list[0]), int(first_list[1]), int(first_list[2]))
        last_date = date(int(last_list[0]), int(last_list[1]), int(last_list[2]))

        return Transaction.objects.filter(
            user=self.request.user, 
            category=category, 
            occur_date__gte=first_date, 
            occur_date__lte=last_date).order_by("-occur_date")


# view to handle the 20 latest transactions of the account
class AccountTransactionList(TransactionView): 

    def get_queryset(self):
        queried_account = get_object_or_404(Account, pk=self.kwargs["pk"])
        return Transaction.objects.filter(account=queried_account).order_by("-occur_date")[:20]
    

"""
Handling the list of latest transactions in each category for the account
requires pagination 
params: category
"""
class AccBothTransactionList(TransactionView): 

    def get_queryset(self):
        arg_category = self.request.query_params.get("category")
        if not arg_category: 
            raise ValidationError({"error": "Category not specified"})
        
        queried_account = get_object_or_404(Account, pk=self.kwargs['pk'])
        
        # the list of transactions with picked category 
        first_date, last_date = get_curr_dates(period_type="month")

        return Transaction.objects.filter(
            account=queried_account, 
            category=arg_category, 
            occur_date__gte=first_date, 
            occur_date__lte=last_date).order_by("-occur_date")