from django.contrib import admin
from .models import (
    BudgetPlan, User, Account, 
    Transaction, Bills, Stock, 
    DateStockPrice, PortfolioValue, OverdueBillMessage
)

admin.site.register(User)
admin.site.register(Account)
admin.site.register(Transaction)
admin.site.register(BudgetPlan)
admin.site.register(Bills)
admin.site.register(Stock)
admin.site.register(DateStockPrice)
admin.site.register(PortfolioValue)
admin.site.register(OverdueBillMessage)