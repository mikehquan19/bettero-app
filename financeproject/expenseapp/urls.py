from django.urls import path 
from . import views
from rest_framework_simplejwt.views import (
    TokenObtainPairView, TokenRefreshView, TokenVerifyView, TokenBlacklistView,
)

urlpatterns = [
    # authentication
    path("login", TokenObtainPairView.as_view()),
    path("login/refresh", TokenRefreshView.as_view()),
    path("login/verify", TokenVerifyView.as_view()),
    path("register", views.Register.as_view()),
    path("logout", TokenBlacklistView.as_view()),
    
    # user's financial summary
    path("summary", views.user_summary_detail),
    path("full_summary", views.user_full_summary_detail),

    # account list, detail, and financial summary
    path("accounts", views.AccountList.as_view()),
    path("accounts/<int:pk>", views.AccountDetail.as_view()),
    path("accounts/<int:pk>/summary", views.AccountSummary.as_view()),

    # transaction list
    path("transactions", views.UserTransactionList.as_view()),

    path("transactions/interval", views.IntervalTransactionList.as_view()),
    path("transactions/category/<str:arg_cat>", views.CategoryTransactionList.as_view()),
    path("transactions/both", views.BothTransactionList.as_view()),

    path("accounts/<int:pk>/transactions", views.AccountTransactionList.as_view()),
    path("accounts/<int:pk>/transactions/both", views.AccBothTransactionList.as_view()),

    # budget list and details 
    path("budget", views.UserBudget.as_view()),
    path("budget/<str:interval_type>", views.UserBudgetDetail.as_view()), 

    # bill's list, details, and overdue messages
    path("bills", views.BillList.as_view()),
    path("bills/<int:pk>", views.BillsDetail.as_view()), 
    path("overdue_message/", views.OverdueMessageList.as_view()),

    # stock list, and details
    path("stocks", views.StockList.as_view()), 
    path("stocks/<str:symbol>", views.StockPriceDetail.as_view()),
    path("portfolio_value", views.PortfolioValueList.as_view())
]