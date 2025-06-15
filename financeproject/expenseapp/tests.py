from .models import (
    category_dict,
    User, Account, Transaction, DateStockPrice, Stock, PortfolioValue
)
from .finance import get_first_and_last_dates
from django.utils import timezone
from django.db import transaction
from django.db.models import F, Sum
from datetime import date, timedelta, datetime
import random

# THE FUNCTIONS TO UPLOAD TEST RECORDS TO THE DATABASE TO TEST FINANCE LOGICS
TEST_ACCOUNT_NUMBERS = [1000, 1001, 2000, 3000] 
AVAILABLE_INSTITUTIONS = ["Chase", "Bank of America", "Wells Fargo", "Citi Bank", "Capital One", "Discover"]

# upload the account 
@transaction.atomic
def upload_accounts(): 
    random.seed(4)
    accounts_to_create = [] 
    for acc_num in TEST_ACCOUNT_NUMBERS: 
        random_institution = random.choice(AVAILABLE_INSTITUTIONS)

        accounts_to_create.append(Account(
            user=User.objects.get(username="mikequan"), 
            account_number=acc_num,
            name=f"Mike Quan's {random_institution}'s Test Account",
            institution=random_institution,
            account_type="Debit", 
            balance=1000
        ))
    
    created_accounts = Account.objects.bulk_create(accounts_to_create)
    print(f"{created_accounts} accounts were created.")


# upload data to database so that we can test utils
@transaction.atomic
def upload_category_transactions(num_transaction: int=3): 
    # create the transactions 
    random.seed(4) # to ensure same sequence of random floats were generated 
    transactions_to_create = []

    for category in list(category_dict.keys()):
        for i in range(num_transaction):
            random_account = Account.objects.get(account_number=random.choice(TEST_ACCOUNT_NUMBERS))

            # append transaction for this month 
            transactions_to_create.append(Transaction(
                user=User.objects.get(username="mikequan"), 
                account=random_account,
                description=f"Test {category} Transaction #{i + 1} this month", 
                category=category,
                amount=round(random.uniform(5, 100), 2), 
                occur_date=timezone.now()
            ))

            # one transaction for the previous month 
            transactions_to_create.append(Transaction(
                user=User.objects.get(username="mikequan"), 
                account=random_account,
                description=f"Test {category} Transaction #{i + 1} previous month", 
                category=category,
                amount=round(random.uniform(5, 100), 2), 
                occur_date=(timezone.now() - timedelta(days=28))
            ))

    created_transactions = Transaction.objects.bulk_create(transactions_to_create)
    print(f"{len(created_transactions)} transactions were created.")


# upload transactions along the interval 
@transaction.atomic
def upload_interval_transactions(num_transactions: int=1): 
    # create transactions 
    random.seed(3)
    transactions_to_create = [] 

    
    current_date = date.today() - timedelta(days=150)
    while current_date <= date.today(): 
        # convert the date to the timezone-aware datetime object 
        converted_date = timezone.make_aware(datetime.combine(current_date, datetime.min.time()))

        # for each date, create the num transactions 
        for i in range(num_transactions): 
            category = random.choice(list(category_dict.keys()))

            # append one transaction for this month 
            transactions_to_create.append(Transaction(
                user=User.objects.get(username="mikequan"), 
                account=Account.objects.get(account_number=random.choice(TEST_ACCOUNT_NUMBERS)),
                description=f"Test {category} Transaction {i + 1}", 
                category=category,
                amount=round(random.uniform(20, 50), 2), 
                occur_date=converted_date
            ))

        current_date += timedelta(days=1)

    created_transactions = Transaction.objects.bulk_create(transactions_to_create)
    print(f"{len(created_transactions)} transactions were created.")


@transaction.atomic
def delete_test_transactions(): 
    Transaction.objects.filter(description__contains="Test").delete()
    print("Test Transaction deleted successfully")


@transaction.atomic
def upload_test_portfolio_values(): 
    user = User.objects.get(username="mikequan")
    stocks = [stock for stock in Stock.objects.filter(user=user)]
    created_portfolio_values = []

    first_date, last_date = get_first_and_last_dates()
    current_date = first_date 
    while current_date <= last_date: 
        date_prices = DateStockPrice.objects.filter(date=current_date, stock__in=stocks)
        date_prices = date_prices.annotate(
            total_value=F("given_date_close") * F("stock__shares"))
        total_value = date_prices.aggregate(total=Sum("total_value", default=0))["total"]

        if total_value != 0: 
            created_portfolio_values.append(
                PortfolioValue(user=user, date=current_date, given_date_value=total_value))
        current_date += timedelta(days=1)

    num_values = PortfolioValue.objects.bulk_create(created_portfolio_values)
    print(f"{len(num_values)} created!")
