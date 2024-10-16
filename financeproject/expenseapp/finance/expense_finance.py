from typing import Dict, List, Tuple
from django.db.models import Sum
from datetime import date, timedelta
from .supplement import *

# THESE ARE FUNCTIONS COMPUTING THE FINANCE OF THE USER'S EXPENSE

# return the total balance of all debit accounts of the user as a tuple 
def total_balance_and_amount_due(arg_user) -> Tuple: 
    # list of debit and credit accounts 
    debit_account_list = arg_user.account_set.filter(account_type="Debit")
    credit_account_list = arg_user.account_set.filter(account_type="Credit")

    # compute the total balance (balance of debit accounts) and amount due (credit accounts)
    total_balance = get_number(debit_account_list.aggregate(total=Sum("balance"))["total"])
    total_amount_due = get_number(credit_account_list.aggregate(total=Sum("balance"))["total"])
    return (total_balance, total_amount_due)


# return the total income of the user of this interval
def total_income(arg_user, arg_first_date=None, arg_last_date=None) -> float: 
    # determine the first and last date of the month
    first_date, last_date = get_current_dates("month", arg_first_date, arg_last_date)

    # query the list of incomes of the user between the first and last date 
    income_list = arg_user.transaction_set.filter(
        from_account=False, 
        occur_date__gte=first_date, 
        occur_date__lte=last_date)
    
    # compute the total income 
    total_income = get_number(income_list.aggregate(total=Sum("amount"))["total"])
    return total_income 


# return the dictionary mapping the date to the total expense of that date 
# up until the current date 
def daily_expense(arg_user, arg_first_date=None, arg_last_date=None) -> Dict: 
    daily_exepense = {} # result dict 
    if not arg_first_date: 
        first_date = date(year=date.today().year, month=date.today().month, day=1)
        last_date = date.today() 
    else: 
        first_date = arg_first_date
        last_date = arg_last_date 
    
    # loop through the dates from first date to today 
    current_date = first_date
    while current_date <= last_date: 
        # query list of expenses between current date and next date 
        current_expense_list = arg_user.transaction_set.filter(
            from_account=True, 
            occur_date__gte=current_date, 
            occur_date__lt=(current_date + timedelta(days=1)))
        # compute the total_expense 
        total_expense = get_number(current_expense_list.aggregate(total=Sum("amount"))["total"])

        # add the mapping betweeen date and expense to the dict
        daily_exepense[current_date.strftime("%m/%d/%Y")] = total_expense
        current_date += timedelta(days=1)
    return daily_exepense 


# calculate the percentage composition of each category of expense
# return the dictionary mapping each category to its composition percentage of this month
def expense_composition_percentage(arg_obj, arg_first_date=None, arg_last_date=None) -> Dict: 
    # the first date and last date of the given month 
    first_date, last_date = get_current_dates("month", arg_first_date, arg_last_date)
    # dictionary mapping the expense's category to the total amount of expense of this month 
    category_expense = category_expense_dict(arg_obj, first_date, last_date)
    # dictionary mapping the expense's category to the percentage of expense
    category_percentage = {"Grocery": 0, "Dining": 0, "Shopping": 0,  
                           "Bills": 0, "Gas": 0, "Others": 0, "Income": 0}

    # total expense indicates that no transactions have been made 
    if category_expense["Total"] != 0:   
        # list of the keys of this dictionary
        for key in list(category_percentage.keys()):  
            category_percentage[key] = (category_expense[key] / category_expense["Total"]) * 100
            # round up to 2 decimal places 
            category_percentage[key] = round(category_percentage[key], 2) 
    return category_percentage


# calculate how the total expenses and expense of each category have changed 
def expense_change_percentage(arg_obj, interval_type="month", arg_first_date=None, arg_last_date=None) -> Dict: 
    # the first and last date of the current month
    curr_first_date, curr_last_date = get_current_dates(interval_type, arg_first_date, arg_last_date)
    # the first and last date of the previous month 
    prev_first_date, prev_last_date = get_previous_dates(interval_type, curr_first_date, curr_last_date)

    # dictionary mapping the expense's category to amount for the current and previous month 
    curr_category_expense = category_expense_dict(arg_obj, curr_first_date, curr_last_date)
    prev_category_expense = category_expense_dict(arg_obj, prev_first_date, prev_last_date)

    # calculate the change percentage 
    # dict mapping each category to the list [current, previous, change percentage]
    change_dict = {"Grocery": 0, "Dining": 0, "Shopping": 0, 
        "Bills": 0, "Gas": 0, "Others": 0, "Income": 0}

    for key in list(change_dict.keys()): 
        if prev_category_expense[key] != 0: 
            # calculate the percentage change and then add to the dict
            # and round to 2 decimal points 
            change_percentage = curr_category_expense[key] - prev_category_expense[key]
            change_percentage = round((change_percentage / prev_category_expense[key]) * 100, 2)
        else: 
            # if no expenses made during previous month, obviously expenses increase 100% 
            if curr_category_expense[key] != 0: 
                change_dict[key]= 100.00
            else: 
                change_dict[key]= 0.00
    return change_dict


# adjust the balance of the debit account based on the amount and flow
# return nothing
def adjust_account_balance(user_account, user_transaction) -> None: 
    # multiplier will determine the result based on if amount is extracted for added
    multiplier = 1 
    if not user_transaction.from_account:
        multiplier = -1
    # if the account is debit, amount is extracted from the balance
    # otherwise, amount is added to the balance
    if user_account.account_type == "Debit":
        user_account.balance -= multiplier * user_transaction.amount
    else:  
        user_account.balance += multiplier * user_transaction.amount
    # save the adjustment to the database 
    user_account.save() 


# return the list of latest intervals (month, bi_week, or week), intervals = tuple (first_date, last_date)
def latest_intervals(interval_type, num_intervals) -> List: 
    # the list of latest time intervals 
    latest_intervals = [] 
    first_date, last_date = get_current_dates(interval_type)
    for _ in range(num_intervals): 
        latest_intervals.append((first_date, last_date))
        first_date, last_date = get_previous_dates(interval_type, first_date, last_date)
    # return the list of intervals 
    return latest_intervals


# return the total expense of each interval depending on the type of the interval
def interval_total_expense(arg_user) -> Dict: 
    # the 5 latest months, bi-weeks, and weeks in the dictionary 
    latest_intervals_dict = {}
    latest_intervals_dict["month"] = latest_intervals("month", 5)
    latest_intervals_dict["bi_week"] = latest_intervals("bi_week", 5)
    latest_intervals_dict["week"] = latest_intervals("week", 5)

    # the dictionary mapping the interval type to the list of expense of each interval
    interval_expense_dict = {}
    for interval_type in list(latest_intervals_dict.keys()): 
        interval_expense_dict[interval_type] = []
        interval_list = latest_intervals_dict[interval_type]

        # compute total expense for each interval of the list 
        for interval in interval_list: 
            # first and last date of the interval 
            first_date = interval[0]
            last_date = interval[1]
            # first key-value pair of total_expense_dict is total expense of all categories
            total_expense = category_expense_dict(arg_user, first_date, last_date)["Total"]

            # the expense change and composition of the user during this interval
            expense_change = expense_change_percentage(arg_user, interval_type, first_date, last_date)
            expense_composition = expense_composition_percentage(arg_user, first_date, last_date)
            
            # daily expense of the user during this interval 
            interval_daily_expense = daily_expense(arg_user, first_date, last_date)
            
            interval_expense_dict[interval_type].append({
                "first_date": first_date, 
                "last_date": last_date, 
                "total_expense": total_expense,
                "expense_change": expense_change, 
                "expense_composition": expense_composition,
                "daily_expense": interval_daily_expense,
            })
    # return the dictionary 
    return interval_expense_dict