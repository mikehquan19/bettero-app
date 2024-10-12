from typing import Dict
from expenseapp.models import BudgetPlan
from .supplement import *

# THESE ARE FUNCTIONS COMPUTING THE FINANCE OF THE USER'S BUDGET AND BILLS

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
