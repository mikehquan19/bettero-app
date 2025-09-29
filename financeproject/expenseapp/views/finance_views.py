from rest_framework import generics
from rest_framework.views import APIView
from rest_framework.permissions import AllowAny, IsAuthenticated
from rest_framework.response import Response
from expenseapp.models import User, PortfolioValue
from django.db import transaction
from expenseapp.serializers import RegisterSerializer
from expenseapp.finance import *

class Register(generics.CreateAPIView): 
    """ Handling the the registration, which is allowed for anyone  """
    permission_classes = [AllowAny]  
    queryset = User.objects.all()
    serializer_class = RegisterSerializer

    def perform_create(self, serializer):
        """
        Override method ```perform_create()``` to create the list of initial portfolio value of the user 
        """
        created_user = serializer.save()
        first_date, last_date = get_first_and_last_dates()
        
        current_date = first_date 
        portfolio_value_list = [] 
        while current_date < last_date: 
            portfolio_value_list.append(PortfolioValue(
                user=created_user, date=current_date, given_date_value=Decimal(0.00)
            ))
            current_date += timedelta(days=1)

        # Ensure the integrity of the query with atomic transaction
        with transaction.atomic():
            PortfolioValue.objects.bulk_create(portfolio_value_list) 
        return created_user
    

class UserSummaryDetail(APIView): 
    """ Handling the info of the financial summary of the user  """
    permission_classes = [IsAuthenticated]

    def get(self, request, format=None) -> Response:
        user = request.user
        first_date, last_date = get_curr_dates("month")
        total_balance, total_amount_due = total_balance_and_amount_due(user)
        financial_info = {
            "total_balance": round(total_balance, 2), 
            "total_amount_due": round(total_amount_due, 2), 
            "total_income": round(total_income(user), 2),
            "total_expense": round(category_expense_dict(user, first_date, last_date)["Total"], 2)
        }

        response_data = { 
            "financial_info": financial_info,
            "change_percentage": expense_change_percentage(user), 
            "composition_percentage": expense_composition_percentage(user), 
            "daily_expense": daily_expense(user), 
        }
        return Response(response_data)
    

class UserFullSummaryDetail(APIView): 
    """ Handling the fully detailed financial summary of the user """
    
    def get(self, request, format=None) -> Response: 
        interval_expense_dict = interval_total_expense(request.user)
        return Response(interval_expense_dict)