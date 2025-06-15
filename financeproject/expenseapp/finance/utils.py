# THESE ARE THE SUPPLEMENTAL FUNCTIONS THAT WILL SUPPORT OTHER FINANCE FUNCTIONS

from ast import Tuple
from typing import Dict
from django.db.models import Sum
from expenseapp.constants import CATEGORY_DICT
from datetime import date, timedelta
from calendar import monthrange

# get the current first and last dates based on the interval type
def get_curr_dates(
    period_type: str=None, arg_first_date: date=None, arg_last_date: date=None
) -> Tuple: 
    
    # Return first and last date if given
    if arg_first_date: return arg_first_date, arg_last_date

    if period_type != "month": 
        # The number of days of an interval (week or bi_week)
        days_between = 7 if period_type == "week" else 14
        # First and last date of the current interval 
        last_date = date.today() + timedelta(days=(6 - date.today().weekday()))
        first_date = last_date - timedelta(days=(days_between - 1))
    else: 
        """
        If the type is month
        first date and last date of the current month
        """
        first_date = date(year=date.today().year, month=date.today().month, day=1)
        last_date = date(
            year=date.today().year, 
            month=date.today().month, 
            day=monthrange(date.today().year, date.today().month)[1]
        )
    return first_date, last_date 


# Return the previous first and last dates 
def get_prev_dates(period_type: str, first_date: date, last_date: date) -> Tuple: 
    if period_type != "month": 
        # The number of days of an interval (week or bi_week)
        days_between = 7 if period_type == "week" else 14
            
        prev_first_date = first_date - timedelta(days=days_between)
        prev_last_date = last_date - timedelta(days=days_between)
    else: 
        curr_year, curr_month = first_date.year, first_date.month
        
        # Compute the previous month, or year
        prev_month, prev_year = curr_month - 1, curr_year
        if prev_month == 0: 
            prev_month, prev_year = 12, curr_year - 1

        prev_first_date = first_date - timedelta(days=monthrange(prev_year, prev_month)[1])
        prev_last_date = last_date - timedelta(days=monthrange(curr_year, curr_month)[1])

    return prev_first_date, prev_last_date 


# get the first date of last month and current date 
def get_first_and_last_dates(): 
    current_date = date.today() # the last date, which is today

    # the first date (which is first date of last month), month and year of the last date 
    prev_month, prev_year = current_date.month - 1, current_date.year
    if prev_month < 0: 
        prev_month, prev_year = 12, prev_year - 1

    first_date_last_month = date(year=prev_year, month=prev_month, day=1)
    return first_date_last_month, current_date


# Return the dict mapping the expense's category to amount 
# between 2 dates
def category_expense_dict(arg_obj, first_date: date, last_date: date) -> Dict:
    
    # Queryset of incomes and expenses between 2 dates 
    expense_list = arg_obj.transaction_set.filter(
        occur_date__gte=first_date, 
        occur_date__lte=last_date).exclude(category="Income")
    
    income_list = arg_obj.transaction_set.filter(
        category="Income",
        occur_date__gte=first_date, occur_date__lte=last_date
    )
    
    """
    Calculate the total expense and each category's expense
    using the GROUP_BY 
    """ 
    category_expense = { category: 0.0 for category in list(CATEGORY_DICT.keys()) }
    annotated_results = expense_list.values("category").annotate(
        total_amount=Sum("amount", default=0)
    ).order_by()
    
    for result in annotated_results: 
        category_expense[result["category"]] = float(result["total_amount"])

    # Compute total expense, and incomes 
    category_expense.update({
        "Expense": float(expense_list.aggregate(total=Sum("amount", default=0))["total"]), 
        "Income": float(income_list.aggregate(total=Sum("amount", default=0))["total"])
    })
    # total transactions is really just sum of expense and income 
    category_expense["Total"] = category_expense["Expense"] + category_expense["Income"]
    return category_expense