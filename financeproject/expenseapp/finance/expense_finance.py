""" THESE ARE FUNCTIONS COMPUTING THE FINANCE OF THE USER'S EXPENSE """

from typing import Dict, List, Tuple
from django.db.models import Sum
from datetime import date, timedelta
from expenseapp.models import Account, Transaction, User
from expenseapp.constants import CATEGORY_DICT
from .utils import *


# Return total debit & credit balance
def total_balance_and_amount_due(user: User) -> Tuple: 
    # List of user's debit and credit accounts 
    debit_list = Account.objects.filter(user=user, account_type="Debit")
    credit_list = Account.objects.filter(user=user, account_type="Credit")

    # Compute sum of each account's list
    total_balance = debit_list.aggregate(total=Sum("balance", default=0))["total"]
    total_credit = credit_list.aggregate(total=Sum("balance", default=0))["total"]

    return total_balance, total_credit


# Return the user's total income of given interval
def total_income(
    arg_user: User, arg_first_date: date=None, arg_last_date: date=None
) -> float: 
    # Determine the first and last date
    first_date, last_date = get_curr_dates("month", arg_first_date, arg_last_date)

    # List of incomes of the user of interval
    income_list = Transaction.objects.filter(
        user=arg_user, category="Income", 
        occur_date__gte=first_date, occur_date__lte=last_date)

    # Compute the total income 
    total_income = income_list.aggregate(total=Sum("amount", default=0))["total"]
    return total_income 


"""
Return the dict mapping date to amount of expense of that date 
up until the current date 
"""
def daily_expense(
    arg_user: User, arg_first_date: date=None, arg_last_date: date=None
) -> Dict: 
    daily_expense = {} # Result dict 

    if not arg_first_date or not arg_last_date: 
        first_date , last_date = date(
            year=date.today().year, month=date.today().month, day=1
        ), date.today()
    else: 
        first_date, last_date = arg_first_date, arg_last_date
    
    # Loop through the dates from first date to today 
    current_date = first_date
    while current_date <= last_date: 
        # Expenses between current date and next date 
        curr_expense_list = Transaction.objects.filter(
            user=arg_user, 
            occur_date__gte=current_date, 
            occur_date__lt=(current_date + timedelta(days=1))
        ).exclude(category="Income")
        
        # Compute the total_expense 
        total_expense = float(curr_expense_list.aggregate(total=Sum("amount", default=0))["total"])

        # Ddd the mapping from date to amount to the dict
        daily_expense[current_date.strftime("%m/%d/%Y")] = total_expense
        current_date += timedelta(days=1)

    return daily_expense 


"""
Compute the percent composition of each expense category
Return the dict mapping category to its composition percentage
""" 
def expense_composition_percentage(
    arg_obj, 
    arg_first_date: date=None, 
    arg_last_date: date=None
) -> Dict: 
    # Dictionary mapping category to total expense this month 
    first_date, last_date = get_curr_dates("month", arg_first_date, arg_last_date)
    expense_dict = category_expense_dict(arg_obj, first_date, last_date)
    
    """
    Dictionary mapping the expense's category to the percentage of expense
    avoid hardcoding the category 
    """
    percent_dict = {
        category : 0.0 for category in list(CATEGORY_DICT.keys()) if category != "Income"
    }
    # Total expense indicates that no transactions have been made 
    if expense_dict["Total"] != 0:  
        # List of the keys of this dictionary is category
        for key in list(percent_dict.keys()):  
            percent_dict[key] = expense_dict[key] / expense_dict["Total"]
            percent_dict[key] = round(percent_dict[key] * 100, 2)

    return percent_dict


# Calculate how the total expenses and expense of each category have changed 
def expense_change_percentage(
    arg_obj, 
    period_type: str="month", 
    arg_first_date: date=None, 
    arg_last_date: date=None
) -> Dict: 

    # First and last date of the current period and the previous period
    curr_date1, curr_date2 = get_curr_dates(period_type, arg_first_date, arg_last_date)
    prev_date1, prev_date2 = get_prev_dates(period_type, curr_date1, curr_date2)

    # Dict mapping category to amount for the current and previous month 
    curr_expense_dict = category_expense_dict(arg_obj, curr_date1, curr_date2)
    prev_expense_dict = category_expense_dict(arg_obj, prev_date1, prev_date2)
 
    # Dict mapping category to the change percentage
    percent_dict = {
        category : 0.0 for category in list(CATEGORY_DICT.keys()) if category != "Income"
    }
    # List of the keys of this dictionary is category
    for key in list(percent_dict.keys()): 
        if prev_expense_dict[key] != 0: 
            # Calculate the percentage change and then add to the dict
            percent_dict[key] = (curr_expense_dict[key] - prev_expense_dict[key]) / prev_expense_dict[key]
            percent_dict[key] = round(percent_dict[key] * 100, 2)
        else: 
            # If no expenses made during previous month, expenses increase 100% 
            percent_dict[key] = 100.00 if curr_expense_dict[key] != 0 else 0.00

    return percent_dict


"""
Adjust the balance of the debit account based on the amount and flow
return nothing
"""
def adjust_account_balance(
    account: Account, transaction: Transaction
) -> None: 
    # determine the result based on if transaction is expense or income
    multiplier = -1 if transaction.category == "Income" else 1

    """
    If the account is debit, amount is extracted from the balance
    otherwise, amount is added to the balance
    """
    if account.account_type == "Debit":
        account.balance -= multiplier * transaction.amount
    else:  
        account.balance += multiplier * transaction.amount
    # Save the adjustment to the database 
    account.save() 


# Return the list of latest intervals (month, bi_week, or week)
# Intervals = tuple (first_date, last_date)
def latest_periods(period_type: str, num_periods: int) -> List: 
    # the list of latest time intervals 
    latest_intervals = [] 
    first_date, last_date = get_curr_dates(period_type)

    for _ in range(num_periods): 
        latest_intervals.append((first_date, last_date))
        first_date, last_date = get_prev_dates(period_type, first_date, last_date)
    return latest_intervals


# Return the total expense of each interval based on the type of the interval
def interval_total_expense(user: User) -> Dict: 
    # Latest months, bi-weeks, and weeks in the dictionary 
    latest_periods_dict = {}
    
    for type in ["month", "bi_week", "week"]: 
        # Get the data of 5 latest periods 
        latest_periods_dict[type] = latest_periods(type, 5)

    # Dict mapping the interval type to the list of expense of each interval
    period_expense_dict = {}
    for period_type in list(latest_periods_dict.keys()): 
        period_expense_dict[period_type] = []
        interval_list = latest_periods_dict[period_type]

        # compute total expense for each interval of the list 
        for interval in interval_list: 
            # First and last date of the interval 
            first_date, last_date = interval[0], interval[1]

            # First dict of total_expense_dict is total expense of all categories
            total_expense = category_expense_dict(user, first_date, last_date)["Total"]

            # The expense change and composition during this period
            expense_change = expense_change_percentage(user, period_type, first_date, last_date)
            expense_composition = expense_composition_percentage(user, first_date, last_date)
            
            # daily expense of the user during this period
            period_daily_expense = daily_expense(user, first_date, last_date)
            
            period_expense_dict[period_type].append({
                "first_date": first_date, 
                "last_date": last_date, 
                "total_expense": total_expense,
                "expense_change": expense_change, 
                "expense_composition": expense_composition,
                "daily_expense": period_daily_expense,
            })
     
    return period_expense_dict