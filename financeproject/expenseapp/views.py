from django.http import Http404
from rest_framework import status, generics
from rest_framework.views import APIView
from rest_framework.permissions import AllowAny, IsAuthenticated
from rest_framework.decorators import api_view, permission_classes
from rest_framework.response import Response
from django.shortcuts import get_object_or_404
from .models import BudgetPlan, User, Account, Transaction, Bills, Stock
from .serializers import (
    BudgetPlanSerializer,
    PortfolioValueSerializer, 
    AccountSerializer,
    RegisterSerializer, 
    TransactionSerializer,
    BillSerializer, 
    StockSerializer,
    StockPriceSerializer, 
    OverdueBillMessageSerializer
)
from . import finance 
from datetime import date, datetime
from calendar import monthrange

class Register(generics.CreateAPIView): 
    permission_classes = [AllowAny] # anyone visiting the page could login 
    queryset = User.objects.all()
    serializer_class = RegisterSerializer


""" VIEWS FOR THE USER'S DETAIL AND SUMMARY INFO"""


# handling the info of the financial summary of the user 
@api_view(["GET"])
@permission_classes([IsAuthenticated])
def user_summary_detail(request): 
    if request.method == "GET": 
        user = request.user

        # first and last dates of month
        first_date, last_date = finance.get_current_dates("mont")

        response_data = {
            # calculate the financial info of the user 
            "total_balance": finance.total_balance_and_amount_due(user)[0], 
            "total_amount_due": finance.total_balance_and_amount_due(user)[1], 
            "total_income": finance.total_income(user), 
            "total_expense": finance.category_expense_dict(user, first_date, last_date)["Total"], 

            # calculate the daily expense, the change, and composition percentage of user 
            "change_percentage": finance.expense_change_percentage(user), 
            "composition_percentage": finance.expense_composition_percentage(user), 
            "daily_expense": finance.daily_expense(user), 
        }
        return Response(response_data)
    else: 
        return Response(
            {"Error": "This API is only for GET method"}, 
            status=status.HTTP_400_BAD_REQUEST)


# handling the fully detailed financial summary of the user 
@api_view(["GET"])
@permission_classes([IsAuthenticated])
def user_full_summary_detail(request): 
    if request.method == "GET": 
        user = request.user
        interval_expense_dict = finance.interval_total_expense(user)

        # initial first date and last date 
        initial_first_date = interval_expense_dict["month"][0]["first_date"]
        initial_last_date = interval_expense_dict["month"][0]["last_date"]

        # compute the initial list of transactions
        initial_transactions = user.transaction_set.filter(
            occur_date__gte=initial_first_date, 
            occur_date__lte=initial_last_date).order_by("-occur_date")
        initial_transaction_data = TransactionSerializer(initial_transactions, many=True).data

        # structure of the response data
        response_data = {
            "latest_interval_expense": interval_expense_dict,
            "initial_transaction_data": initial_transaction_data,
        }
        return Response(response_data)
    else: 
        return Response(
            {"Error": "This API is only for GET method"}, 
            status=status.HTTP_400_BAD_REQUEST)
    

""" VIEWS FOR THE ACCOUNT LIST, DETAIL, AND SUMMARY INFO """


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

    # POST method, create new account to list of users
    def post(self, request, format=None):
        request_data = request.data 
        request_data["user"] = request.user.id
        new_account_serializer = AccountSerializer(data=request_data)
        if new_account_serializer.is_valid(raise_exception=True): 
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
        change_percentage = finance.expense_change_percentage(account)
        composition_percentage = finance.expense_composition_percentage(account)

        # the response data 
        response_data = {
            "change_percentage": change_percentage, 
            "composition_percentage": composition_percentage, 
        }
        return Response(response_data)
    else: 
        return Response(
            {"Error": "This API is only for GET method"}, 
            status=status.HTTP_400_BAD_REQUEST)


""" VIEWS FOR THE TRANSACTIONS LIST """


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


# handling the list of transactions between 2 predefined dates 
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
        first_date, last_date = finance.get_current_dates(arg_interval_type="month")

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

""" VIEWS FOR THE OTHER BUDGET PLAN AND BILLS """

# handling the budget plan of the user 
class UserBudget(APIView): 
    permission_classes = [IsAuthenticated]

    def get_response_data(self, request): 
        # query and serialize the user 
        user = request.user
        custom_data = {"month": {}, "bi_week": {}, "week": {}}
        for key in list(custom_data.keys()): 
            custom_data[key] = finance.get_budget_response_data(user, key)
        return custom_data 
    
    # GET method 
    def get(self, request, format=None): 
        response_data = self.get_response_data(request)
        return Response(response_data)
    
    # POST method 
    def post(self, request, format=None): 
        request_data = request.data
        request_data["user"] = request.user.id
        new_plan_serializer = BudgetPlanSerializer(data=request_data)

        if new_plan_serializer.is_valid(raise_exception=True): 
            new_plan_serializer.save() # call the create method 

            # return new budget plan
            response_data = self.get_response_data(request)
            return Response(response_data, status=status.HTTP_201_CREATED)
        return Response(new_plan_serializer.errors, status=status.HTTP_400_BAD_REQUEST)
    

# handling the the budget plan of each interval type 
class UserBudgetDetail(APIView): 
    permission_classes = [IsAuthenticated]

    def get_budget_plan(self, request, interval_type): 
        # query the user and check if the user has plan of this type
        user = request.user
        try: 
            return user.budgetplan_set.get(interval_type=interval_type)
        except BudgetPlan.DoesNotExist: 
            raise Http404("Budget plan doesn't exist!")
    
    # GET method, just return the plan of the given interval type
    def get(self, request, interval_type, format=None): 
        user = request.user
        response_data = finance.get_budget_response_data(user, interval_type)
        return Response(response_data)
    
    # PUT method, update the plan of the given interval type 
    def put(self, request, interval_type, format=None): 
        request_data = request.data
        request_data["user"] = request.user.id
        budget_plan = self.get_budget_plan(request, interval_type)
        updated_plan_serializer = BudgetPlanSerializer(budget_plan, data=request_data)
    
        if updated_plan_serializer.is_valid(raise_exception=True): 
            updated_plan_serializer.save() # call update() method 
            # custom data to be returned 
            custom_data = {"month": {}, "bi_week": {}, "week": {}}
            for key in list(custom_data.keys()): 
                custom_data[key] = finance.get_budget_response_data(request.user, key)
            return Response(custom_data, status=status.HTTP_202_ACCEPTED)
        return Response(updated_plan_serializer.errors, status=status.HTTP_400_BAD_REQUEST)
    
    # DELETE method, delete the plan 
    def delete(self, request, interval_type, format=None): 
        budget_plan = self.get_budget_plan(request, interval_type)
        budget_plan.delete()
        return Response(status=status.HTTP_204_NO_CONTENT)


# handling the list of bills of the user 
class BillsList(APIView): 
    # get the response data 
    def get_response_data(self, request):
        user = request.user
        bills_list = user.bills_set.all()
        response_data = BillSerializer(bills_list, many=True).data
        return response_data 
    
    # GET method, return list of bills for the user 
    def get(self, request, format=None): 
        response_data = self.get_response_data(request)
        return Response(response_data)
    
    # POST method, add new bill to the list of bills 
    def post(self, request, format=None): 
        request_data =request.data
        request_data["user"] = request.user.id

        new_bill_serializer = BillSerializer(data=request_data)
        if new_bill_serializer.is_valid(raise_exception=True): 
            new_bill_serializer.save() # call the create() method 
            # return the new list of bills 
            response_data = self.get_response_data(request)
            return Response(response_data, status=status.HTTP_201_CREATED)
        return Response(new_bill_serializer.errors, status=status.HTTP_400_BAD_REQUEST)


# handling the detail of the bills 
class BillsDetail(generics.RetrieveUpdateDestroyAPIView):
    permission_classes = [IsAuthenticated]
    serializer_class = BillSerializer

    # get object and the serializer class 
    def get_object(self):
        user = self.request.user
        try: 
            return user.bills_set.get(pk=self.kwargs["pk"])
        except Bills.DoesNotExist: 
            raise Http404

    # overriding the destroying behavior 
    def perform_destroy(self, instance):
        # if there is pay account and the bills isn't overdue yet
        if instance.pay_account != None and instance.due_date >= date.today():
            # create transactions contending that we've paid the bills 
            # before destroying the bills 
            new_transaction = Transaction.objects.create(
                user=instance.user, 
                account=instance.pay_account, 
                description=f"Payment: {instance.description}",
                amount=instance.amount,
                from_account=True,
                occur_date=datetime.now(),
                category="Bills")
            # adjust the account 
            finance.adjust_account_balance(
                user_account=new_transaction.account, 
                user_transaction=new_transaction)
        # destroy the bills 
        instance.delete()


@api_view(["GET"])
@permission_classes([IsAuthenticated])
def overdue_message_list(request): 
    if request.method == "GET": 
        user = request.user
        overdue_message_list = user.overduebillmessage_set.all()
        response_data = OverdueBillMessageSerializer(
            overdue_message_list, 
            many=True).data
        return Response(response_data)
    else: 
        return Response(
            {"Error": "This API is only for GET method"}, 
            status=status.HTTP_400_BAD_REQUEST)
    

""" VIEWS FOR THE STOCK'S VALUE AND STUFF """


# handling the list of stocks 
class StockList(APIView): 
    permission_classes = [IsAuthenticated]

    # get the response data 
    def get_response_data(self, request):
        user = request.user
        stock_list = user.stock_set.all()
        response_data = StockSerializer(stock_list, many=True).data
        return response_data
    
    # GET method, return the list of stock of the user 
    def get(self, request, format=None): 
        response_data = self.get_response_data(request)
        return Response(response_data)
    
    # POST method, add the stock the list of stock of user 
    def post(self, request, format=None): 
        request_data = request.data 
        request_data["user"] = request.user.id
        symbol = request_data["symbol"]

        # load the price data of the stock with the given symbol 
        try: 
            stock_data = finance.load_stock_data(symbol)
        except IndexError: # if the stock with the given symbol isn't found 
            return Response(
                {"error": "No stock with the given symbol"}, 
                status=status.HTTP_404_NOT_FOUND)
        
        stock_price_data = stock_data.pop("price_data")
        for key in list(stock_data.keys()): 
            request_data[key] = stock_data[key]

        # add to the list 
        new_stock_serializer = StockSerializer(data=request_data)
        if new_stock_serializer.is_valid(raise_exception=True): 
            created_stock = new_stock_serializer.save() 

            # add all of the price of the stock
            for i in range(len(stock_price_data)): 
                created_stock.datestockprice_set.create(
                    date=stock_price_data[i]["date"],
                    given_date_close=stock_price_data[i]["given_date_close"])

            # return the response data
            response_data = self.get_response_data(request)
            return Response(response_data, status=status.HTTP_201_CREATED)
        return Response(new_stock_serializer.errors, status=status.HTTP_400_BAD_REQUEST)
    

# handling the price detail of the stock 
class StockPriceDetail(APIView): 
    permission_classes = [IsAuthenticated]

    def get_response_data(self, request, symbol): 
        user = request.user
        stock = get_object_or_404(Stock, user=user, symbol=symbol)
        # list of prices of the stock 
        stock_price_list = stock.datestockprice_set.order_by("date")

        # response data 
        response_data = {}
        response_data["stock"] = StockSerializer(stock).data
        response_data["price_list"] = {}

        price_list = StockPriceSerializer(stock_price_list, many=True).data
        for price in price_list: 
            response_data["price_list"][price["date"]] = price["given_date_close"]
        return response_data
    
    # GET method, return the detail of the stock, including its list of price
    def get(self, request, symbol, format=None): 
        response_data = self.get_response_data(request, symbol)
        return Response(response_data)
    
    # PUT method, update the stock 
    def put(self, request, symbol, format=None): 
        user = request.user
        request_data = request.data
        request_data["user"] = user.id
        stock = get_object_or_404(Stock, user=user, symbol=symbol)

        updated_stock_serializer = StockSerializer(stock, data=request_data)
        if updated_stock_serializer.is_valid(raise_exception=True): 
            updated_stock_serializer.save() 
            # return the response data
            response_data = self.get_response_data(request, symbol)
            return Response(response_data, status=status.HTTP_202_ACCEPTED)
        return Response(updated_stock_serializer.errors, status=status.HTTP_400_BAD_REQUEST)
    
    # DELETE method, delete the stock 
    def delete(self, request, symbol, format=None): 
        user = request.user
        stock = get_object_or_404(Stock, user=user, symbol=symbol)
        stock.delete()
        return Response({"message": "Stock deleted successfully"})


@api_view(["GET"])
@permission_classes([IsAuthenticated])
def portfolio_value_list(request): 
    if request.method == "GET": 
        # query the list of portfolios
        user = request.user
        portfolio_value_list = user.portfoliovalue_set.order_by("date")
        original_data = PortfolioValueSerializer(portfolio_value_list, many=True).data 

        response_data = {}
        for data_item in original_data: 
            response_data[data_item["date"]] = data_item["given_date_value"]
        return Response(response_data)
    else: 
        return Response(
            {"Error": "This API is only for GET method"}, 
            status=status.HTTP_400_BAD_REQUEST)
