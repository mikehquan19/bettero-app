from typing import Dict
from django.db.models import Sum
from datetime import date, timedelta
from calendar import monthrange

# THESE ARE THE SUPPLEMENTAL FUNCTIONS THAT WILL SUPPORT OTHER FINANCE FUNCTIONS
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
    
    # compute the sum of all transactions, expense transactions, and income transactions
    category_expense["Total"] = get_number(transaction_list.aggregate(total=Sum("amount"))["total"])
    category_expense["Expense"] = get_number(expense_list.aggregate(total=Sum("amount"))["total"])
    category_expense["Income"] = get_number(income_list.aggregate(total=Sum("amount"))["total"])
    # compute the total expense of each category 
    for expense in expense_list: 
        category_expense[expense.category] += expense.amount 
    return category_expense