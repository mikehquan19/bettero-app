from django.http import Http404
from django.shortcuts import get_object_or_404
from rest_framework import status, generics
from rest_framework.views import APIView
from rest_framework.permissions import IsAuthenticated
from rest_framework.response import Response
from expenseapp.models import BudgetPlan, OverdueBillMessage, Transaction, Bill
from expenseapp.serializers import BudgetPlanSerializer, BillSerializer, OverdueBillMessageSerializer
from django.db import transaction

from expenseapp.finance import get_budget_response_data, adjust_account_balance
from datetime import date, datetime

class UserBudget(APIView):
    """ View to handle the budget plan of the user  """

    permission_classes = [IsAuthenticated]

    def get_response_data(self, request): 
        response_data = {
            "month": {}, 
            "bi_week": {}, 
            "week": {}
        }
        for type in list(response_data.keys()): 
            response_data[type] = get_budget_response_data(arg_user=request.user, period_type=type)
        return response_data
    
    def get(self, request, format=None) -> Response: 
        response_data = self.get_response_data(request)
        return Response(response_data)
    
    def post(self, request, format=None) -> Response: 
        request_data = {
            **request.data, 
            "user": request.user.id
        }

        new_plan_serializer = BudgetPlanSerializer(data=request_data)
        if new_plan_serializer.is_valid(): 
            new_plan_serializer.save() 
            response_data = self.get_response_data(request)
            return Response(response_data, status=status.HTTP_201_CREATED)
        
        return Response(new_plan_serializer.errors, status=status.HTTP_400_BAD_REQUEST)
    

class UserBudgetDetail(APIView):
    """ View of handle the budget plan of the user each interval type  """

    permission_classes = [IsAuthenticated]

    def get_budget_plan(self, request, interval_type: str): 
        """ query the user and check if the user has plan of this type """
        try: 
            queried_plan = request.user.budgetplan_set.get(interval_type=interval_type)
        except BudgetPlan.DoesNotExist: 
            raise Http404("Budget plan with given type doesn't exist.")
        
        return queried_plan
    
    def get(self, request, interval_type: str, format=None) -> Response: 
        """ GET method, return the plan of the given interval type """

        response_data = get_budget_response_data(request.user, interval_type)
        return Response(response_data)
     
    def put(self, request, interval_type: str, format=None) -> Response: 
        """ PUT method, update the plan of the given interval type """
        request_data = {
            **request.data, "user": request.user.id
        }
        
        plan = self.get_budget_plan(request, interval_type)
        updated_plan_serializer = BudgetPlanSerializer(plan, data=request_data)
    
        if updated_plan_serializer.is_valid(): 
            updated_plan_serializer.save() # call update() method 

            # custom data to be returned 
            response_data = {
                "month": {}, 
                "bi_week": {}, 
                "week": {}
            }

            for type in list(response_data.keys()): 
                response_data[type] = get_budget_response_data(request.user, type)
            return Response(response_data, status=status.HTTP_202_ACCEPTED)
        
        return Response(updated_plan_serializer.errors, status=status.HTTP_400_BAD_REQUEST)
    
    def delete(self, request, interval_type, format=None) -> Response: 
        """ DELETE method, delete the plan  """

        budget_plan = self.get_budget_plan(request, interval_type)
        budget_plan.delete()
        return Response(status=status.HTTP_204_NO_CONTENT)


class BillList(APIView): 
    """ View to handle the list of bills of the user """
 
    def get_response_data(self, request):
        bills_list = Bill.objects.filter(user=request.user)
        response_data = BillSerializer(bills_list, many=True).data
        return response_data 
    
    def get(self, request, format=None) -> Response: 
        """ GET method, return list of bills for the user """
        response_data = self.get_response_data(request)
        return Response(response_data)
    
    def post(self, request, format=None) -> Response: 
        """ POST method, add new bill to the list of bills  """
        request_data = { **request.data, "user": request.user.id }
        new_bill_serializer = BillSerializer(data=request_data)
        if new_bill_serializer.is_valid(): 
            new_bill_serializer.save() 

            # return the new list of bills 
            response_data = self.get_response_data(request)
            return Response(response_data, status=status.HTTP_201_CREATED)
        
        return Response(new_bill_serializer.errors, status=status.HTTP_400_BAD_REQUEST)


class BillsDetail(generics.RetrieveUpdateDestroyAPIView):
    """ View to handle the detail of the bills  """

    permission_classes = [IsAuthenticated]
    serializer_class = BillSerializer

    def get_object(self):
        """ Override ```get_object(self)``` to get from pk by the endpoint"""
        return get_object_or_404(Bill, id=self.kwargs["pk"])

    @transaction.atomic
    def perform_destroy(self, instance):
        """ Override the destroying behavior """
        
        if instance.pay_account is not None and instance.due_date >= date.today():
            # if there is pay account and the bills isn't overdue yet
            # create transactions indicating that user's paid the bills 
            new_transaction = Transaction.objects.create(
                account=instance.pay_account, user=instance.user, 
                description=f"Payment: {instance.description}", category=instance.category,
                amount=instance.amount, occur_date=datetime.now()
            )
            # Adjust the account 
            adjust_account_balance(new_transaction.account, new_transaction)

        # destroy the bills 
        instance.delete()


class OverdueMessageList(APIView): 
    permission_classes = [IsAuthenticated]

    def get(self, request, format=None) -> Response: 
        """ GET method """
        overdue_message_list = OverdueBillMessage.objects.filter(user=request.user)
        response_data = OverdueBillMessageSerializer(overdue_message_list, many=True).data
        return Response(response_data)
    