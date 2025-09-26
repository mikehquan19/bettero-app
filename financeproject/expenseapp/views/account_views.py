from django.db import transaction
from django.shortcuts import get_object_or_404
from rest_framework import status, generics
from rest_framework.views import APIView
from rest_framework.permissions import IsAuthenticated
from rest_framework.response import Response
from expenseapp.models import Account, Transaction
from expenseapp.serializers import AccountSerializer
from expenseapp.finance import expense_change_percentage, expense_composition_percentage
import datetime

from expenseapp.constants import CREDIT, OTHERS, INCOME

class AccountList(APIView):   
    """ View to handle the list of accounts of the specific user  """

    permission_classes = [IsAuthenticated]
    
    def get_response_data(self, request):
        """ Return the customized response data with the user id """

        # Query and serialize the account list 
        account_list = Account.objects.filter(user=request.user)
        response_data = AccountSerializer(account_list, many=True).data
        return response_data

    def get(self, request, format=None) -> Response:
        """ GET method, return the list of accounts of the user """

        response_data = self.get_response_data(request)
        return Response(response_data)

    def post(self, request, format=None) -> Response:
        """ POST method, create new account to list of accounts and then return them """

        request_data = request.data 
        request_data["user"] = request.user.pk
        
        new_account_serializer = AccountSerializer(data=request_data)
        if new_account_serializer.is_valid(): 
            new_account_serializer.save() # Call the method create()

            # Return the new list of accounts
            response_data = self.get_response_data(request)
            return Response(response_data, status=status.HTTP_201_CREATED)
        
        return Response(new_account_serializer.errors, status=status.HTTP_400_BAD_REQUEST)
    

class AccountDetail(generics.RetrieveUpdateDestroyAPIView): 
    """ View to handle the detail of the account """

    permission_classes = [IsAuthenticated]
    serializer_class = AccountSerializer
    queryset = Account.objects.all()
    
    @transaction.atomic
    def perform_update(self, serializer) -> None:
        """ 
        Override ```perform_update``` to add the new transaction showing change of balance.
        Helping with the consistency of the word
        """
        
        previous_balance = self.get_object().balance
        updated_account = serializer.save()
        current_balance = updated_account.balance
        balance_change = current_balance - previous_balance

        # Create the transaction corresponding to the change, if any.
        if balance_change == 0: return

        # The transaction's description differs based on the change
        if balance_change > 0: 
            description = f"Balance increases ${abs(balance_change)}"
            category = OTHERS if updated_account.account_type == CREDIT else INCOME
        else: 
            description = f"Balance decreases ${abs(balance_change)}"
            category = INCOME if updated_account.account_type == CREDIT else OTHERS
                
        # Create transaction to fill in the discrepancy in the account's balance
        Transaction.objects.create(
            user=updated_account.user, account=updated_account, description=description, 
            amount=abs(balance_change), occur_date=datetime.datetime.now(), 
            category=category
        )


class AccountSummary(APIView): 
    """ View to handle the info of the financial summary of the specific account """

    permission_classes = [IsAuthenticated]

    def get(self, request, pk, format=None) -> None: 
        queried_account = get_object_or_404(Account, pk=pk)
        
        # Get the change & composition percentage of the account 
        change_percentage = expense_change_percentage(queried_account)
        composition_percentage = expense_composition_percentage(queried_account)

        response_data = {
            "change_percentage": change_percentage, 
            "composition_percentage": composition_percentage, 
        }
        return Response(response_data)