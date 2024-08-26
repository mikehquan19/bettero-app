from decimal import Decimal
from typing import Dict, List, Tuple
from django.shortcuts import get_object_or_404
from django.db.models import Sum
from .models import BudgetPlan
from datetime import date, timedelta
from calendar import monthrange
import yfinance as yf


# turn None into 0
def get_number(arg): 
    return 0 if arg == None else arg

# get the current first and last dates 
def get_current_dates(arg_interval_type=None, arg_first_date=None, arg_last_date=None): 
    if arg_first_date == None: 
        if arg_interval_type != "month": 
            # the number of days of the intervals of given type
            in_between_days = 7 # for week
            if arg_interval_type == "bi_week": 
                in_between_days = 14 
            # first and last date of the current interval 
            last_date = date.today() + timedelta(days=(6 - date.today().weekday()))
            first_date = last_date - timedelta(days=(in_between_days - 1))
        # if the type is month
        else: 
            # first date and last date of the current month
            first_date = date(year=date.today().year, month=date.today().month, day=1)
            last_date = date(
                year=date.today().year, 
                month=date.today().month, 
                day=monthrange(date.today().year, date.today().month)[1])
    else: 
        first_date = arg_first_date
        last_date = arg_last_date 
    return first_date, last_date 


# get the previous first and last dates 
def get_previous_dates(arg_interval_type, arg_first_date, arg_last_date): 
    if arg_interval_type != "month": 
        # the number of days of the intervals of given type
        in_between_days = 7 # for week
        if arg_interval_type == "bi_week": 
            in_between_days = 14 
            
        prev_first_date = arg_first_date - timedelta(days=in_between_days)
        prev_last_date = arg_last_date - timedelta(days=in_between_days)
    else: 
        curr_year = arg_first_date.year
        curr_month = arg_first_date.month
        
        # compute the previous month, and year (if necessary)
        prev_month = curr_month - 1
        prev_year = curr_year
        if prev_month == 0:
            prev_month = 12
            prev_year = curr_year - 1

        prev_first_date = arg_first_date - timedelta(days=monthrange(prev_year, prev_month)[1])
        prev_last_date = arg_last_date - timedelta(days=monthrange(curr_year, curr_month)[1])

    # return the first date, last date
    return prev_first_date, prev_last_date 


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


# return the dictionary mapping the expense's category to amount for the interval between 2 dates
def category_expense_dict(arg_obj, arg_first_date, arg_last_date) -> Dict:
    # query the list of transactions (in general) incomes and expenses between 2 dates 
    transaction_list = arg_obj.transaction_set.filter(
        occur_date__gte=arg_first_date, 
        occur_date__lte=arg_last_date)
    expense_list = transaction_list.filter(from_account=True)
    income_list = transaction_list.filter(from_account=False)
    
    # calculate the total expense and the expense of each category
    # dictionary mapping the expense's category to amount for the interval between 2 dates to be returned 
    category_expense = {"Total": 0, "Expense": 0 ,"Grocery": 0, "Dining": 0, "Shopping": 0, 
                        "Bills": 0, "Gas": 0, "Others": 0, "Income": 0}
    
    # compute the sum of all transactions 
    category_expense["Total"] = get_number(transaction_list.aggregate(total=Sum("amount"))["total"])
    category_expense["Expense"] = get_number(expense_list.aggregate(total=Sum("amount"))["total"])
    category_expense["Income"] = get_number(income_list.aggregate(total=Sum("amount"))["total"])

    
    # compute the total expense of each category 
    for expense in expense_list: 
        category_expense[expense.category] += expense.amount 
    return category_expense


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
            change_percentage = curr_category_expense[key] - prev_category_expense[key]
            change_percentage = (change_percentage / prev_category_expense[key]) * 100
            # round the change to 2
            change_dict[key]= round(change_percentage, 2) 
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


# calculate the actual compostion percentage of each category vs the goal
def budget_composition_percentage(arg_user, interval_type) -> Dict: 
    # query the budget plan 
    plan = arg_user.budgetplan_set.get(interval_type=interval_type)

    # the first and last date of the current interval of given type 
    first_date, last_date = get_current_dates(interval_type)

    # the total expense of each category 
    category_expenses = category_expense_dict(arg_user, first_date, last_date)

    # dictionary mapping the type's name to the set of composition percentage of that type
    budget_percentage = {
        "goal": {
            "Grocery": plan.grocery, 
            "Dining": plan.dining, 
            "Shopping": plan.shopping,
            "Bills": plan.bills, 
            "Gas": plan.gas, 
            "Others": plan.others,
        },
        "actual": {}
    }
    if category_expenses["Expense"] != 0: 
        for key in list(budget_percentage["goal"].keys()): 
            # calculate the percentage and round up to 2 decimal places 
            budget_percentage["actual"][key] = (category_expenses[key] / category_expenses["Expense"]) * 100
            budget_percentage["actual"][key] = round(budget_percentage["actual"][key], 2)
    return budget_percentage


# calculate the the progress percentage of each towards that category's budget
def budget_progress_percentage(arg_user, interval_type) -> Dict: 
    # query the budget plan and total budget
    plan = arg_user.budgetplan_set.get(interval_type=interval_type)
    total_budget = plan.recurring_income * plan.portion_for_expense / 100

    # the first and last date of the current interval of given type 
    first_date, last_date = get_current_dates(interval_type)

    # the total expense of each category 
    category_expenses = category_expense_dict(arg_user, first_date, last_date)

    # dictionary mapping each category to its current expense,
    # budget, and its progress percentage 
    progress_percentage = {
        "Expense": {"budget": total_budget}, 
        "Grocery": {"budget": plan.grocery * total_budget / 100}, 
        "Dining": {"budget": plan.dining * total_budget / 100}, 
        "Shopping": {"budget": plan.shopping * total_budget / 100}, 
        "Bills": {"budget": plan.bills * total_budget / 100}, 
        "Gas": {"budget": plan.gas * total_budget / 100}, 
        "Others": {"budget": plan.others * total_budget / 100}, 
    }

    # iterate through each key in the progress percentage dict
    for key in list(progress_percentage.keys()):
        # the current expense of this category 
        progress_percentage[key]["current"] = category_expenses[key]

        # if the budget is not $0, calculate the percentage 
        if progress_percentage[key]["current"] < progress_percentage[key]["budget"]: 
            # the progress percentage 
            percentage = (progress_percentage[key]["current"] / progress_percentage[key]["budget"]) * 100
            progress_percentage[key]["percentage"] = round(percentage, 2)
        # otherwise, percentage is automatically 100%
        else: 
            progress_percentage[key]["percentage"] = 100
    return progress_percentage


# get response data for 
def get_budget_response_data(arg_user, type) -> Dict: 
    # data about the budget 
    budget_response = {}
    try: 
        this_plan = arg_user.budgetplan_set.get(interval_type=type)
    except BudgetPlan.DoesNotExist: 
        return budget_response
    
    # id and the income  
    budget_response["id"] = this_plan.id
    budget_response["income"] = this_plan.recurring_income

    # the budget percentage and total budget 
    budget_response["expense_portion"] = this_plan.portion_for_expense

    # composition percentage and progress percentage 
    budget_composition_dict = budget_composition_percentage(arg_user, type)
    budget_progress_dict = budget_progress_percentage(arg_user, type)
    budget_response["composition"] = budget_composition_dict
    budget_response["progress"] = budget_progress_dict
    return budget_response


# get the first date of last month and current date 
def get_first_and_last_dates(): 
    # the last date (which is today)
    current_date = date.today()

    # the first date (which is first date of last month)
    # month and year of the last date 
    current_month = current_date.month 
    current_year = current_date.year 

    prev_month = current_month - 1
    prev_year = current_year
    if prev_month < 0: 
        prev_month = 12
        prev_year -= 1

    first_date_last_month = date(year=prev_year, month=prev_month, day=1)
    return first_date_last_month, current_date


# convert the date obj to string 
def to_string(arg_date): 
    return f"{arg_date.year}-{arg_date.month}-{arg_date.day}"


# convert string to the date obj
def to_date(arg_str): 
    str_arr = arg_str.split("-")
    return date(year=int(str_arr[0]), month=int(str_arr[1]), day=int(str_arr[2]))


# load the initial price of the stock since first date of last month
# till the today along with current data of the stock 
def load_stock_data(symbol) -> Dict: 
    # get the first and last date 
    first_date, last_date = get_first_and_last_dates()

    # load data of the stock's info 
    recent_data = yf.download([symbol], start=to_string(first_date), end=to_string(last_date))

    # structure of the data 
    custom_data = {}

    # current info of the stock 
    custom_data["current_close"] = round(Decimal(recent_data["Adj Close"].iloc[-1]), 2)
    custom_data["previous_close"] = round(Decimal(recent_data["Adj Close"].iloc[-2]), 2)
    custom_data["open"] = round(Decimal(recent_data["Open"].iloc[-1]), 2)
    custom_data["high"] = round(Decimal(recent_data["High"].iloc[-1]), 2)
    custom_data["low"] = round(Decimal(recent_data["Low"].iloc[-1]), 2)
    custom_data["volume"] = int(recent_data["Volume"].iloc[-1])

    # the price of the stock over the past
    custom_data["price_data"] = []
    current_date = first_date 
    while current_date < last_date: 
        try: 
            given_date_price = recent_data["Adj Close"][to_string(current_date)]
            item_data = {
                "date": current_date, 
                "given_date_close": float(round(given_date_price, 2))
            }
            custom_data["price_data"].append(item_data)
        # the key error means that the price of stock on that date doesn't exist
        except KeyError: 
            pass
        # increment the date
        current_date += timedelta(days=1)

    # the date the price of the stock was updated 
    last_updated_date = custom_data["price_data"][-1]["date"]
    custom_data["last_updated_date"] = last_updated_date
    return custom_data


# update the info the stock, and add new record of the stock price 
def update_stock_data(symbol) -> Dict: 
    previous_date = to_string(date.today() - timedelta(days=1))
    today = to_string(date.today())
    
    # return the Panda data frame 
    updated_data = yf.download(symbol, start=previous_date, end=today)
    
    custom_data = {}
    custom_data["new_close"] = round(Decimal(updated_data["Adj Close"].iloc[0]), 2)
    custom_data["new_open"] = round(Decimal(updated_data["Open"].iloc[0]), 2)
    custom_data["new_high"] = round(Decimal(updated_data["High"].iloc[0]), 2)
    custom_data["new_low"] = round(Decimal(updated_data["Low"].iloc[0]), 2)
    custom_data["new_volume"] = int(updated_data["Volume"].iloc[0])
    return custom_data