from decimal import Decimal
from django.db import models
from django.contrib.auth.models import AbstractUser
from django.core.exceptions import ValidationError
from django.core.validators import MinValueValidator, MaxValueValidator
from django.utils import timezone

# the user account of the app 
class User(AbstractUser): 
    # login info
    full_name = models.CharField(max_length=50)
    user_email = models.EmailField()
    username = models.CharField(max_length=20, default="username", unique=True)
    password = models.CharField(max_length=20, default="password")

    # date joined will be automatically be today
    date_joined = models.DateTimeField(default=timezone.now)

    # representation of the user (using their name)
    def __str__(self):
        return self.username; 


# the financial accounts of the user, in general 
class Account(models.Model): 
    user = models.ForeignKey(User, on_delete=models.CASCADE, default=1)
    account_number = models.IntegerField()
    name = models.CharField(max_length=50)
    institution = models.CharField(max_length=50)
    account_type = models.CharField(
        max_length=20, 
        choices={
            "Debit": "DEBIT", 
            "Credit": "CREDIT"
        })
    balance = models.DecimalField(
        max_digits=10, 
        decimal_places=2, 
        default=0, 
        validators=[MinValueValidator(limit_value=0.0)])
    # this is strictly for credit accounts 
    credit_limit = models.DecimalField(max_digits=10, decimal_places=2, null=True, blank=True)
    due_date = models.DateField(null=True, blank=True)

    # representation of the account using the account's name 
    def __str__(self): 
        return self.name
    

# transactions in general of the accounts 
class Transaction(models.Model): 
    # user and account associated with it 
    account = models.ForeignKey(Account, on_delete=models.CASCADE, default=1)
    user = models.ForeignKey(User, on_delete=models.CASCADE, default=1)

    # info the accounts 
    description = models.CharField(max_length=200)
    amount = models.DecimalField(
        max_digits=10, 
        decimal_places=2, 
        default = 0, 
        validators=[MinValueValidator(limit_value=0.01)])
    from_account = models.BooleanField(default=True)
    # hours were used to sort 
    occur_date = models.DateTimeField("The date transaction was made")
    category = models.CharField(
        max_length=30, 
        choices={
            "Grocery": "GROCERY", 
            "Dining": "FOOD & DRINK", 
            "Shopping": "SHOPPING", 
            "Bills": "MONTHLY BILLS",
            "Gas": "GAS",
            "Others": "OTHERS",
        }
    )

    # representation of the transaction using its content message
    def __str__(self): 
        return self.description
    

# the budget plan of the user over the 
class BudgetPlan(models.Model): 
    user = models.ForeignKey(User, on_delete=models.CASCADE, default=1)
    # type of the budget plan (budget over month, bi-week, or week)
    interval_type = models.CharField(
        max_length=20,
        choices={
            "month": "MONTH", 
            "bi_week": "BI WEEK", 
            "week": "WEEK", 
        },
    )

    # the recurring income over that period (in dollars)
    recurring_income = models.DecimalField(
        max_digits=10, 
        decimal_places=2, 
        default=0.01,
        validators=[MinValueValidator(limit_value=0.01)])
    
    # validator for the percentage 
    value_validators = [
        MinValueValidator(limit_value=0),
        MaxValueValidator(limit_value=100), 
    ]
    # the portion of the recurring income used for expense (in percentage)
    portion_for_expense = models.DecimalField(max_digits=5, decimal_places=2, default=0, validators=value_validators)

    # the portion (based on the amount of money for expense) on each category
    grocery = models.DecimalField(max_digits=5, decimal_places=2, default=0, validators=value_validators)
    dining = models.DecimalField(max_digits=5, decimal_places=2, default=0, validators=value_validators)
    shopping = models.DecimalField(max_digits=5, decimal_places=2, default=0, validators=value_validators)
    bills = models.DecimalField(max_digits=5, decimal_places=2, default=0, validators=value_validators)
    gas = models.DecimalField(max_digits=5, decimal_places=2, default=0, validators=value_validators)
    others = models.DecimalField(max_digits=5, decimal_places=2, default=0, validators=value_validators)

    class Meta: 
        # this will group the budget plans of same user together 
        ordering = ["user"]

    # validate the budget plan 
    def clean(self): 
        # the user this budget plans belongs to 
        user = self.user

        # if the user doesn't have any budget plan, skip it 
        if user.budgetplan_set.exists(): 
            # the queryset of user's budget plan, excluding this instance
            budget_plan_list = user.budgetplan_set.exclude(id=self.id)
            # the maximum number of budget plans, excluding this instance 
            max_plan_count = 2 

            # validate if the number of budget plans exceed 3 
            if budget_plan_list.count() > max_plan_count: 
                raise ValidationError("The number of budget plans must be less than or equal 3")

            # validate if this budget plan overlaps the previous ones
            budget_plan_list = budget_plan_list.filter(interval_type=self.interval_type)
            if budget_plan_list.exists(): 
                raise ValidationError(f"This budget plan's interval overlaps the previous ones ({self.id})")

        # validate if the category percentages don't add up to 100%
        if self.grocery + self.dining + self.shopping + self.bills + self.gas + self.others != 100: 
            raise ValidationError("All of the category portions don't add up to 100%")

    # override in order to invoke clean() method before saving the instance 
    def save(self, *args, **kwargs): 
        self.full_clean()
        return super().save(*args, **kwargs)

    # representation of the budget plan using the plan's user name and its type
    def __str__(self): 
        return f"{self.user}'s {self.interval_type} budget plan"


# the montly bills of the user 
class Bills(models.Model): 
    user = models.ForeignKey(User, on_delete=models.CASCADE, default=1)
    pay_account = models.ForeignKey(Account, null=True, blank=True, on_delete=models.SET_NULL, default=1)
    description = models.CharField(max_length=200)
    amount = models.DecimalField(
        max_digits=10, 
        decimal_places=2, 
        default = 0, 
        validators=[MinValueValidator(limit_value=Decimal(1.00))])
    due_date = models.DateField() 

    def __str__(self): 
        return self.description
    

# the value of the user's portfolio
class PortfolioValue(models.Model): 
    user = models.ForeignKey(User, on_delete=models.CASCADE, default=1)
    date = models.DateField()

    # valuer of the porfolio on the given date 
    min_validator = [MinValueValidator(limit_value=Decimal(0.00))]
    given_date_value = models.DecimalField(max_digits=10, decimal_places=2, default=0, validators=min_validator)

    def __str__(self): 
        return f"{self.user}'s portfolio value on {self.date}"


# the stock that the user holds
class Stock(models.Model): 
    user = models.ForeignKey(User, on_delete=models.CASCADE, default=1)
    corporation = models.CharField(max_length=100)
    name = models.CharField(max_length=100)
    symbol = models.CharField(max_length=10)
     
    min_validator = [MinValueValidator(limit_value=Decimal(0.00))]
    shares = models.DecimalField(max_digits=10, decimal_places=2, default=0, validators=min_validator)

    # close price of the stock on the updated date 
    previous_close = models.DecimalField(max_digits=10, decimal_places=2, default=0, validators=min_validator)
    current_close = models.DecimalField(max_digits=10, decimal_places=2, default=0, validators=min_validator)

    # open, low, and high price of the stock 
    open = models.DecimalField(max_digits=10, decimal_places=2, default=0, validators=min_validator)
    low = models.DecimalField(max_digits=10, decimal_places=2, default=0, validators=min_validator)
    high = models.DecimalField(max_digits=10, decimal_places=2, default=0, validators=min_validator)

    # volume of the stock's trading 
    volume = models.BigIntegerField(default=0)

    # the last time the stock's close was updated 
    last_updated_date = models.DateField("The last date the stock was updated")

    # validate if the user has 2 different stocks of the same symbol 
    def clean(self): 
        # the user this stock belongs to 
        user = self.user
        if user.stock_set.exists(): 
            # the list of stock with this stock symbol, but excluding this stock
            stock_list = user.stock_set.exclude(id=self.id)
            stock_list = stock_list.filter(symbol=self.symbol)
            if stock_list.exists(): 
                raise ValidationError(f"This stock overlaps the previous ones ({self.id})")
    
     # override in order to invoke clean() method before saving the instance 
    def save(self, *args, **kwargs): 
        self.full_clean()
        return super().save(*args, **kwargs) 

    # representation of the stock 
    def __str__(self): 
        return self.symbol
    

# the price of the stock of the specific date 
# only store the price of the stock on any date as of the first date of last month 
# (1 month + x days of this month)
class DateStockPrice(models.Model): 
    stock = models.ForeignKey(Stock, on_delete=models.CASCADE, default=1)
    date = models.DateField()

    # the close price of the given stock on the given date 
    given_date_close =  models.DecimalField(
        max_digits=10, 
        decimal_places=2, 
        default=0,
        validators=[MinValueValidator(limit_value=Decimal(0.00))])
    
    class Meta: 
        # order the stock price based on the stock and the date 
        ordering = ["stock", "date"]
    
    # representation of the stock's price 
    def __str__(self): 
        return f"{self.stock.symbol}'s close on {self.date}"
    

# the message telling the user that the there are overdue bills 
class OverdueBillMessage(models.Model): 
    user = models.ForeignKey(User, on_delete=models.CASCADE, default=1)
    bill_description = models.CharField(max_length=200)
    bill_amount = models.DecimalField(
        max_digits=10, 
        decimal_places=2, 
        default = 0, 
        validators=[MinValueValidator(limit_value=Decimal(1.00))])
    bill_due_date = models.DateField("The date the overdue bill was due",null=True)
    appear_date = models.DateField("The date the message was created", null=True)

    def __str__(self): 
        return self.bill_description
    
