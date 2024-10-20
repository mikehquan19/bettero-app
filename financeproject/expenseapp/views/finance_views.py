from rest_framework import generics
from rest_framework.permissions import AllowAny, IsAuthenticated
from rest_framework.decorators import api_view, permission_classes
from rest_framework.response import Response
from expenseapp.models import User, Transaction
from expenseapp.serializers import RegisterSerializer, TransactionSerializer
from expenseapp.finance import *

# handling the the registration, which is allowed for anyone 
class Register(generics.CreateAPIView): 
    permission_classes = [AllowAny] # anyone visiting the page could login 
    queryset = User.objects.all()
    serializer_class = RegisterSerializer


# handling the info of the financial summary of the user 
@api_view(["GET"])
@permission_classes([IsAuthenticated])
def user_summary_detail(request): 
    if request.method == "GET": 
        queried_user = request.user
        # first and last dates of month
        first_date, last_date = get_current_dates("month")

        response_data = {
            # calculate the financial info of the user 
            "total_balance": total_balance_and_amount_due(queried_user)[0], 
            "total_amount_due": total_balance_and_amount_due(queried_user)[1], 
            "total_income": total_income(queried_user), 
            "total_expense": category_expense_dict(queried_user, first_date, last_date)["Total"], 

            # calculate the daily expense, the change, and composition percentage of user 
            "change_percentage": expense_change_percentage(queried_user), 
            "composition_percentage": expense_composition_percentage(queried_user), 
            "daily_expense": daily_expense(queried_user), 
        }
        return Response(response_data)


# handling the fully detailed financial summary of the user 
@api_view(["GET"])
@permission_classes([IsAuthenticated])
def user_full_summary_detail(request): 
    if request.method == "GET": 
        queried_user = request.user
        interval_expense_dict = interval_total_expense(queried_user)

        # initial first date and last date 
        initial_first_date = interval_expense_dict["month"][0]["first_date"]
        initial_last_date = interval_expense_dict["month"][0]["last_date"]

        # compute the initial list of transactions
        initial_transactions = Transaction.objects.filter(
            user=queried_user,
            occur_date__gte=initial_first_date, 
            occur_date__lte=initial_last_date).order_by("-occur_date")
        initial_transaction_data = TransactionSerializer(initial_transactions, many=True).data

        # structure of the response data
        response_data = {
            "latest_interval_expense": interval_expense_dict,
            "initial_transaction_data": initial_transaction_data,
        }
        return Response(response_data)