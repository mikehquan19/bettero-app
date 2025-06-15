""" THESE ARE FUNCTIONS COMPUTING THE FINANCE OF THE USER'S BUDGET AND BILLS """

from typing import Dict
from expenseapp.models import BudgetPlan, User
from .utils import *


# Calculate the actual compostion percentage of each category vs the goal
def budget_composition_percentage(arg_user: User, period_type: str) -> Dict: 
    plan = BudgetPlan.objects.get(user=arg_user, interval_type=period_type)
 
    # The total expense of each category 
    first_date, last_date = get_curr_dates(period_type)
    category_expense = category_expense_dict(arg_user, first_date, last_date)

    # Dict mapping the type to the set of composition percentage of that type
    percent_dict = {
        "goal": {}, 
        "actual": {}
    }

    # Each key representing category
    for key in list(plan.category_portion.keys()): 
        percent_dict["goal"][key] = plan.category_portion[key]

    # If the total expense is 0, there is no composition
    total_expense = category_expense["Expense"]
    if total_expense == 0: return percent_dict

    for key in list(percent_dict["goal"].keys()): 
        this_category_expense = category_expense[key]
        # Calculate the percentage  
        percent_dict["actual"][key] = (this_category_expense / total_expense) * 100
        percent_dict["actual"][key] = round(percent_dict["actual"][key], 2)

    return percent_dict


# Calculate the the progress percentage of each towards that category's budget
def budget_progress_percentage(arg_user: User, interval_type: str) -> Dict: 

    # Budget plan and total budget
    plan = BudgetPlan.objects.get(user=arg_user, interval_type=interval_type)
    total_budget = float(plan.recurring_income * plan.portion_for_expense / 100)
 
    """
    The total expense of each category 
    dates of the current interval of given type
    """
    first_date, last_date = get_curr_dates(interval_type) 
    category_expense = category_expense_dict(arg_user, first_date, last_date)

    # Dictionary mapping each category to its current expense, budget, and its progress percentage 
    percent_dict = {
        "Expense": {
            "budget": total_budget
        }
    }

    # Add budget to the category first
    for key in list(plan.category_portion.keys()):
        this_category_budget = float(plan.category_portion[key])
        percent_dict[key] = { "budget": this_category_budget * total_budget / 100 }

    # Iterate through each key in the progress percentage dict to add current
    for key in list(percent_dict.keys()):
        # the current expense of this category 
        percent_dict[key]["current"] = category_expense[key]

        # make the code look more neat, really 
        current_progress = percent_dict[key]["current"]
        budget_progress = percent_dict[key]["budget"]

        # calculate the progress percentage 
        if current_progress < budget_progress: 
            percentage = (current_progress / budget_progress) * 100
            percent_dict[key]["percentage"] = round(percentage, 2)
        else: 
            # otherwise, percentage is automatically 100%
            percent_dict[key]["percentage"] = 100

    return percent_dict


# get response data for budget plan
def get_budget_response_data(arg_user: User, period_type: str) -> Dict: 
    # Budget data
    budget_response = {}
    try: 
        plan = BudgetPlan.objects.get(user=arg_user, interval_type=period_type)
    except BudgetPlan.DoesNotExist: 
        return budget_response # Return the empty dictionary 
    
    # composition percentage and progress percentage
    composition_dict = budget_composition_percentage(arg_user, period_type)
    progress_dict = budget_progress_percentage(arg_user, period_type)

    budget_response = {
        "id": plan.pk,   
        "income": plan.recurring_income,
        "expense_portion": plan.portion_for_expense, # the budget percentage vs total budget
        "composition": composition_dict, 
        "progress": progress_dict
    }
    
    return budget_response
