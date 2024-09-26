from django.http import Http404
from rest_framework import status
from rest_framework.views import APIView
from rest_framework.permissions import IsAuthenticated
from rest_framework.decorators import api_view, permission_classes
from rest_framework.response import Response
from expenseapp.models import Account, Transaction
from expenseapp.serializers import TransactionSerializer
from expenseapp.finance import get_current_dates
from datetime import date
from calendar import monthrange


# handling the list of transactions of the user 
class UserTransactionList(APIView): 
    permission_classes = [IsAuthenticated]

    # get custom data as a response to the API 
    def get_response_data(self, request):
        # query and serialize the transaction list 
        user = request.user
        transaction_list = user.transaction_set.order_by("-occur_date")[:15]
        response_data = TransactionSerializer(transaction_list, many=True).data
        return response_data
    
    # GET method, return the list of 15 latest transactions 
    def get(self, request, format=None): 
        response_data = self.get_response_data(request)
        return Response(response_data)
    
    # POST method, create new transaction
    def post(self, request, format=None): 
        new_trans_serializer = TransactionSerializer(data=request.data)
        if new_trans_serializer.is_valid(): 
            new_trans_serializer.save() # call the create method 

            response_data = self.get_response_data(request)
            return Response(response_data, status=status.HTTP_201_CREATED)
        return Response(new_trans_serializer.errors, status = status.HTTP_400_BAD_REQUEST)


# handling the list of transactions between 2 given dates in the endpoints 
@api_view(["GET"])
@permission_classes([IsAuthenticated])
def user_interval_transactions(request, arg_first_date, arg_last_date): 
    if request.method == "GET": 
        user = request.user

        # first and last date of the interval 
        first_list = arg_first_date.split("-")
        last_list= arg_last_date.split("-")
        first_date = date(int(first_list[0]), int(first_list[1]), int(first_list[2]))
        last_date = date(int(last_list[0]), int(last_list[1]), int(last_list[2]))

        # the list of transactions between 2 predefined dates 
        transaction_list = user.transaction_set.filter(
            occur_date__gte=first_date, 
            occur_date__lte=last_date).order_by("-occur_date")
        response_data = TransactionSerializer(transaction_list, many=True).data
        return Response(response_data)
    else: 
        return Response(
            {"Error": "This API is only for GET method"}, 
            status=status.HTTP_400_BAD_REQUEST)


# handling the list of latest transactions in each category for the user 
@api_view(["GET"])
@permission_classes([IsAuthenticated])
def user_category_transactions(request, arg_category): 
    if request.method == "GET": 
        user = request.user

        # the list of transactions with picked category
        first_date = date(year=date.today().year, month=date.today().month, day=1)
        last_date = date(
            year=date.today().year, 
            month=date.today().month, 
            day=monthrange(date.today().year, date.today().month)[1])
        
        # data about the list of transactions with the picked category 
        from_account = True if arg_category != "Income" else False
        category = arg_category if arg_category != "Income" else "Others"
        transaction_list = user.transaction_set.filter(
            from_account=from_account, 
            category=category, 
            occur_date__gte=first_date, 
            occur_date__lte=last_date,
        ).order_by("-occur_date")

        response_data = TransactionSerializer(transaction_list, many=True).data
        return Response(response_data)
    else: 
        return Response(
            {"Error": "This API is only for GET method"}, 
            status=status.HTTP_400_BAD_REQUEST)


# handling the list of transactions of the given category between 2 predefined dates 
@api_view(["GET"])
@permission_classes([IsAuthenticated])
def interval_category_transactions(request, arg_category, arg_first_date, arg_last_date):
    if request.method == "GET":  
        user = request.user

        # first and last date of the interval 
        first_list = arg_first_date.split("-")
        last_list = arg_last_date.split("-")
        first_date = date(int(first_list[0]), int(first_list[1]), int(first_list[2]))
        last_date = date(int(last_list[0]), int(last_list[1]), int(last_list[2]))
        
        # the list of transactions with the picked category 
        from_account = True if arg_category != "Income" else False
        category = arg_category if arg_category != "Income" else "Others"
        transaction_list = user.transaction_set.filter(
            from_account=from_account,
            category=category, 
            occur_date__gte=first_date, 
            occur_date__lte=last_date
        ).order_by("-occur_date")
        
        response_data = TransactionSerializer(transaction_list, many=True).data
        return Response(response_data)
    else: 
        return Response(
            {"Error": "This API is only for GET method"}, 
            status=status.HTTP_400_BAD_REQUEST)


# GET method, return the 15 latest transactions of the account
@api_view(["GET"])
@permission_classes([IsAuthenticated])
def account_transaction_list(request, pk):  
    if request.method == "GET": 
        user = request.user
        try: 
            account = user.account_set.get(pk=pk)
        except Account.DoesNotExist: 
            raise Http404
        
        # the trasaction list of the given account 
        transaction_list = account.transaction_set.order_by("-occur_date")[:15]
        response_data = TransactionSerializer(transaction_list, many=True).data
        return Response(response_data)
    else: 
        return Response(
            {"Error": "This API is only for GET method"}, 
            status=status.HTTP_400_BAD_REQUEST)


# handling the list of latest transactions in each category for the account 
@api_view(["GET"])
@permission_classes([IsAuthenticated])
def account_category_transactions(request, pk, arg_category): 
    if request.method == "GET": 
        user = request.user
        try: 
            account = user.account_set.get(pk=pk)
        except Account.DoesNotExist: 
            raise Http404

        # the list of transactions with picked category
        first_date, last_date = get_current_dates(arg_interval_type="month")

        # the list of transactions with the picked category 
        from_account = True if arg_category != "Income" else False
        category = arg_category if arg_category != "Income" else "Others"
        transaction_list = account.transaction_set.filter(
            from_account=from_account,
            category=category, 
            occur_date__gte=first_date, 
            occur_date__lte=last_date
        ).order_by("-occur_date")
            
        response_data = TransactionSerializer(transaction_list, many=True).data
        return Response(response_data)
    else: 
        return Response(
            {"Error": "This API is only for GET method"}, 
            status=status.HTTP_400_BAD_REQUEST)