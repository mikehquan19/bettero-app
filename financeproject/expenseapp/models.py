from decimal import Decimal
from bcrypt import kdf
from django.db import models
from django.contrib.auth.models import AbstractUser
from django.core.exceptions import ValidationError
from django.core.validators import MinValueValidator, MaxValueValidator
from .constants import CATEGORY_DICT

def get_default_dict(): 
    return dict(
        housing=10, 
        automobile=10, 
        medical=10, 
        subscription=10, 
        grocery=10, 
        dining=10, 
        shopping=10, 
        gas=10, 
        others=20
    )

class User(AbstractUser):
    """The user account of the app"""
    def __str__(self):
        return self.username; 


class Account(models.Model): 
    """The financial account of the user, in general"""
    user = models.ForeignKey(User, on_delete=models.CASCADE, default=1)
    account_number = models.IntegerField(unique=True)
    name = models.CharField(max_length=50)
    institution = models.CharField(max_length=50, 
        choices={
            "JP Morgan Chase": "JP MORGAN CHASE",
            "Bank of America": "BANK OF AMERICA",
            "Wells Fargo": "WELLS FARGO",
            "Citi Bank": "CITI BANK",
            "Capital One": "CAPITAL ONE",
            "Discover": "DISCOVER",
            "Sofi Bank": "SOFI BANK",
            "Ally Bank": "ALLY BANK"
        }
    )
    type = models.CharField(max_length=20, choices={
        "Debit": "DEBIT", 
        "Credit": "CREDIT"
    })
    balance = models.DecimalField(
        max_digits=12, decimal_places=2, default=0, 
        validators=[
            MinValueValidator(limit_value=Decimal(0.0))
        ]
    )
    # For credit accounts 
    credit_limit = models.DecimalField(max_digits=12, decimal_places=2, null=True, blank=True)
    due_date = models.DateField(null=True, blank=True)

    def __str__(self): 
        return self.account_number


class Transaction(models.Model):
    """Transaction in general of the account"""
    account = models.ForeignKey(Account, on_delete=models.CASCADE, default=1)
    user = models.ForeignKey(User, on_delete=models.CASCADE, default=1)
    description = models.CharField(max_length=200)
    category = models.CharField(max_length=30, choices=CATEGORY_DICT)
    amount = models.DecimalField(
        max_digits=10, decimal_places=2, default = 0, 
        validators=[
            MinValueValidator(limit_value=Decimal(0.01))
        ]
    )
    occur_date = models.DateTimeField(db_index=True)

    def __str__(self): 
        return self.description


class BudgetPlan(models.Model): 
    """The budget plan of the user"""
    user = models.ForeignKey(User, on_delete=models.CASCADE, default=1)
    # type of the budget plan (budget over month, bi-week, or week)
    interval_type = models.CharField(max_length=20, choices={ 
        "month": "MONTH", 
        "bi_week": "BI WEEK", 
        "week": "WEEK"
    })
    recurring_income = models.DecimalField(
        max_digits=10, decimal_places=2, default=Decimal(0.01), 
        validators=[
            MinValueValidator(limit_value=Decimal(0.01))
        ]
    )
    # Portion of the recurring income for expense in percentage 
    portion_for_expense = models.DecimalField(
        max_digits=5, decimal_places=2, default=Decimal(0.0), 
        validators=[
            MinValueValidator(limit_value=Decimal(0)),
            MaxValueValidator(limit_value=Decimal(100)), 
        ])
    category_portion = models.JSONField(default=get_default_dict)

    class Meta: 
        # Group the budget plans of same user together 
        ordering = ["user"]

    def clean(self): 
        """Custom validate the budget plan"""
        # Queryset of user's budget plan, excluding this instance. 
        # If no instance, id = None
        budget_plans = self.objects.filter(user=self.user).exclude(id=self.id)

        # Validate if the number of budget plans exceed 3 
        if budget_plans.count() >= 3: 
            raise ValidationError("There must be less than or equal 3 plans")

        # Validate if this budget plan overlaps the previous ones
        budget_plans = budget_plans.filter(
            interval_type=self.interval_type
        )
        if budget_plans.exists(): 
            raise ValidationError(f"Plan ({self.id})'s interval overlaps the previous ones")

        # Validate if the category percentages don't add up to 100%
        total = sum([float(percent) for percent in self.category_portion.values()])
        if total != 100: 
            raise ValidationError("All of the category portions don't add up to 100%")
 
    def save(self, *args, **kwargs): 
        """Override to invoke clean() method saving the instance"""
        self.full_clean()
        return super().save(*args, **kwargs)

    def __str__(self): 
        return f"{self.user}'s {self.interval_type}"

 
class Bill(models.Model): 
    """The monthly bills of the user"""
    user = models.ForeignKey(User, on_delete=models.CASCADE, default=1)
    account = models.ForeignKey(
        Account, null=True, blank=True, on_delete=models.SET_NULL, default=1
    )
    description = models.CharField(max_length=200)
    category = models.CharField(max_length=30, choices=CATEGORY_DICT, default="Housing")
    amount = models.DecimalField(
        max_digits=10, decimal_places=2, default = 0, 
        validators=[MinValueValidator(limit_value=Decimal(1.00))])
    due_date = models.DateField() 

    def __str__(self): 
        return self.description
    

class PortfolioValue(models.Model): 
    """The value of the user's portfolio on given date"""
    user = models.ForeignKey(User, on_delete=models.CASCADE, default=1)
    date = models.DateField()
    value = models.DecimalField(
        max_digits=10, decimal_places=2, default=0, 
        validators=[
            MinValueValidator(limit_value=Decimal(0.00))
        ]
    )

    def __str__(self): 
        return f"{self.user}'s portfolio value on {self.date}"


class Stock(models.Model):
    """The stock that the user holds"""
    user = models.ForeignKey(User, on_delete=models.CASCADE, default=1)
    corporation = models.CharField(max_length=100)
    name = models.CharField(max_length=100)
    symbol = models.CharField(max_length=10)
    shares = models.DecimalField(
        max_digits=10, decimal_places=2, default=0, 
        validators=[
            MinValueValidator(limit_value=Decimal(0.00))
        ]
    )

    # close price of the stock on the updated date 
    previous_close = models.DecimalField(
        max_digits=10, decimal_places=2, default=0, 
        validators=[
            MinValueValidator(limit_value=Decimal(0.00))
        ]
    )
    current_close = models.DecimalField(
        max_digits=10, decimal_places=2, default=0, 
        validators=[
            MinValueValidator(limit_value=Decimal(0.00))
        ]
    )

    # open, low, and high price of the stock 
    open = models.DecimalField(
        max_digits=10, decimal_places=2, default=0, 
        validators=[
            MinValueValidator(limit_value=Decimal(0.00))
        ]
    )
    low = models.DecimalField(
        max_digits=10, decimal_places=2, default=0, 
        validators=[
            MinValueValidator(limit_value=Decimal(0.00))
        ]
    )
    high = models.DecimalField(
        max_digits=10, decimal_places=2, default=0, 
        validators=[
            MinValueValidator(limit_value=Decimal(0.00))
        ]
    )
    volume = models.BigIntegerField(default=0)
    updated_date = models.DateField()

    def clean(self): 
        """Validate if the user has 2 different stocks of the same symbol"""
        stock_list = self.objects.exclude(
            id=self.id
        ).filter(user=self.user, symbol=self.symbol)
        if stock_list.exists(): 
            raise ValidationError(f"Stock ({self.id}) overlaps the previous one")
    
    def save(self, *args, **kwargs): 
        """Override in invoke clean() method before saving the instance"""
        self.full_clean()
        return super().save(*args, **kwargs) 
 
    def __str__(self): 
        return self.symbol
    

class DateStockPrice(models.Model): 
    """
    The price of the stock of the specific date, 
    only store the price of the stock on any date as of the first date of last month 
    (1 month & number of days of this month)
    """
    stock = models.ForeignKey(Stock, on_delete=models.CASCADE, default=1)
    date = models.DateField()
    close =  models.DecimalField(
        max_digits=10, decimal_places=2, default=0, 
        validators=[MinValueValidator(limit_value=Decimal(0.00))]
    )
    
    class Meta:  
        # Order the stock price based on the stock and the date
        ordering = ["stock", "date"]
    
    def __str__(self): 
        return f"{self.stock.symbol}'s close on {self.date}"
    

class OverdueBillMessage(models.Model): 
    """The message telling the user that the there are overdue bills"""
    user = models.ForeignKey(User, on_delete=models.CASCADE, default=1)
    description = models.CharField(max_length=200)
    amount = models.DecimalField(
        max_digits=10, 
        decimal_places=2, 
        default = 0, 
        validators=[MinValueValidator(limit_value=Decimal(1.00))]
    )
    due_date = models.DateField(null=True)
    appear_date = models.DateField(null=True)

    def __str__(self): 
        return self.bill_description
    
