from django.http import Http404
from rest_framework import status, generics
from rest_framework.views import APIView
from rest_framework.permissions import IsAuthenticated
from rest_framework.decorators import api_view, permission_classes
from rest_framework.response import Response
from expenseapp.models import Account
from expenseapp.serializers import AccountSerializer
from expenseapp.finance import (
    expense_change_percentage, 
    expense_composition_percentage
)

# handling the list of accounts of the user 
class AccountList(APIView):   
    permission_classes = [IsAuthenticated]
    # get the customized response data with the user id
    def get_response_data(self, request):
        # query and serialize the account list 
        user = request.user
        account_list = user.account_set.all()
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
        user = self.request.user
        try: 
            selected_account =  user.account_set.get(pk=self.kwargs["pk"])
        except Account.DoesNotExist: 
            raise Http404
        return selected_account
    

# handling the info of the financial summary of the specific account
@api_view(["GET"])
@permission_classes([IsAuthenticated])
def account_summary_detail(request, pk):
    if request.method == "GET": 
        user = request.user
        try: 
            account = user.account_set.get(pk=pk)
        except Account.DoesNotExist: 
            return Response({"Error": "Account not found"}, status=status.HTTP_404_NOT_FOUND)
        
        # calculate the change & composition percentage of the account 
        change_percentage = expense_change_percentage(account)
        composition_percentage = expense_composition_percentage(account)

        # the response data 
        response_data = {
            "change_percentage": change_percentage, 
            "composition_percentage": composition_percentage, 
        }
        return Response(response_data)