from django.http import Http404
from django.db import transaction
from rest_framework import status, generics
from rest_framework.views import APIView
from rest_framework.permissions import IsAuthenticated
from rest_framework.decorators import api_view, permission_classes
from rest_framework.response import Response
from expenseapp.models import Account, Transaction
from expenseapp.serializers import AccountSerializer
from expenseapp.finance import (
    expense_change_percentage, 
    expense_composition_percentage
)
import datetime

# handling the list of accounts of the user 
class AccountList(APIView):   
    permission_classes = [IsAuthenticated]
    # get the customized response data with the user id
    def get_response_data(self, request):
        # query and serialize the account list 
        queried_user = request.user
        account_list = Account.objects.filter(user=queried_user)
        response_data = AccountSerializer(account_list, many=True).data
        # return the json
        return response_data

    # GET method, return the list of accounts of the user
    def get(self, request, format=None):
        response_data = self.get_response_data(request)
        return Response(response_data)

    # POST method, create new account to list of accounts and then return them 
    def post(self, request, format=None):
        request_data = request.data 
        request_data["user"] = request.user.id
        new_account_serializer = AccountSerializer(data=request_data)
        if new_account_serializer.is_valid(): 
            # call the create method 
            new_account_serializer.save() 

            # return the new list of accounts
            response_data = self.get_response_data(request)
            return Response(response_data, status=status.HTTP_201_CREATED)
        return Response(new_account_serializer.errors, status=status.HTTP_400_BAD_REQUEST)
    

# handling the detail of account 
class AccountDetail(generics.RetrieveUpdateDestroyAPIView): 
    permission_classes = [IsAuthenticated]
    serializer_class = AccountSerializer

    # the account instance 
    def get_object(self): 
        try: 
            selected_account =  Account.objects.get(pk=self.kwargs["pk"])
        except Account.DoesNotExist: 
            raise Http404("Account with the given pk not found.")
        return selected_account
    
    @transaction.atomic
    def perform_update(self, serializer):
        previous_balance = self.get_object().balance
        updated_account = serializer.save()

        current_balance = updated_account.balance
        balance_change = current_balance - previous_balance

        # create the transaction corresponding to the change, if any
        if balance_change != 0: 
            # The content of the transaction differs depending on change
            if balance_change > 0: 
                description = f"Account's balance increases ${abs(balance_change)}"
                from_account = True if updated_account.account_type == "Credit" else False
            else: 
                description = f"Account's balance decreases ${abs(balance_change)}"
                from_account = False if updated_account.account_type == "Credit" else True
            # create transaction
            Transaction.objects.create(
                user=updated_account.user, account=updated_account, description=description, 
                amount=abs(balance_change), from_account=from_account, 
                occur_date=datetime.datetime.now(), category="Others",
            )

    

# handling the info of the financial summary of the specific account
@api_view(["GET"])
@permission_classes([IsAuthenticated])
def account_summary_detail(request, pk):
    if request.method == "GET": 
        try: 
            queried_account = Account.objects.get(pk=pk)
        except Account.DoesNotExist: 
            raise Http404("Account with the given pk not found.")
        
        # calculate the change & composition percentage of the account 
        change_percentage = expense_change_percentage(queried_account)
        composition_percentage = expense_composition_percentage(queried_account)

        # the response data 
        response_data = {
            "change_percentage": change_percentage, 
            "composition_percentage": composition_percentage, 
        }
        return Response(response_data)