from django.urls import path 
from . import views
from rest_framework_simplejwt.views import (
    TokenObtainPairView,
    TokenRefreshView,
    TokenVerifyView, 
    TokenBlacklistView,
)

app_name = "expenseapp" # creating the namespace for the app
urlpatterns = [
    # Authentication
    path("login/", TokenObtainPairView.as_view(), name="login"),
    path("login/refresh/", TokenRefreshView.as_view(), name="login_refresh"),
    path("login/verify/", TokenVerifyView.as_view(), name="login_verify"),
    path("register/", views.Register.as_view(), name="register"),
    path("logout/", TokenBlacklistView.as_view(), name="logout"),
    
    # user's summary
    path("summary/", views.user_summary_detail, name="user_summary_detail"),
    path("full_summary/", views.user_full_summary_detail, name="user_full_summary_detail"),

    # account list, detail, and summary
    path("accounts/", views.AccountList.as_view(), name="account_list"),
    path("accounts/<int:pk>/", views.AccountDetail.as_view(), name="account_detail"),
    path("accounts/<int:pk>/summary/", views.account_summary_detail, name="account_summary_detail"),

    # transaction list, and summary
    path("transactions/", views.UserTransactionList.as_view(), name="user_transaction_list"),
    path(
        "transactions/<str:arg_first_date>/to/<str:arg_last_date>/",
        views.user_interval_transactions, 
        name="user_interval_transactions"),
    path("transactions/category/<str:arg_category>/", views.user_category_transactions, name="user_category_transactions"),
    path(
        "transactions/category/<str:arg_category>/<str:arg_first_date>/to/<str:arg_last_date>/",
        views.interval_category_transactions,
        name="interval_category_transaction"),

    path("accounts/<int:pk>/transactions/", views.account_transaction_list, name="account_transaction_list"),

    path(
        "accounts/<int:pk>/transactions/category/<str:arg_category>/", 
        views.account_category_transactions, 
        name="account_category_transactions"),

    # budget list and details 
    path("budget/", views.UserBudget.as_view(), name="user_budget"),
    path("budget/<str:interval_type>/", views.UserBudgetDetail.as_view(), name="user_budget_detail"), 
    path("bills/", views.BillsList.as_view(), name="bills_list"),

    # bill's list, details, and overdue messages
    path("bills/<int:pk>/", views.BillsDetail.as_view(), name="bills_detail"), 
    path("overdue_message/", views.overdue_message_list, name="overdue_message_list"),

    # stock list, and details
    path("stocks/", views.StockList.as_view(), name="stock_list"), 
    path("stocks/<str:symbol>/", views.StockPriceDetail.as_view(), name="stock_price_detail"),
    path("portfolio_value/", views.portfolio_value_list, name="porfolio_value_list")
    
]