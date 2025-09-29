""" THESE ARE FUNCTIONS COMPUTING THE FINANCE OF THE USER'S BUDGET AND BILLS """

from typing import Dict

from django.shortcuts import get_object_or_404
from expenseapp.models import BudgetPlan, User
from .utils import get_curr_dates, category_expense_dict


def budget_composition_percentage(user: User, interval_type: str) -> Dict: 
    """ Calculate the actual compostion percentage of each category along with the goal """

    budget_plan = BudgetPlan.objects.get(user=user, interval_type=interval_type)
    # The total expense of each category 
    first_date, last_date = get_curr_dates(interval_type)
    category_expense = category_expense_dict(user, first_date, last_date)

    # Dict mapping the type to the set of composition percentage of that type
    percent_dict = {
        "goal": {}, 
        "actual": {}
    }

    # Each key representing category
    for category in list(budget_plan.category_portion.keys()): 
        percent_dict["goal"][category] = budget_plan.category_portion[category]

    # If the total expense is 0, there is no composition
    total_expense = category_expense["Expense"]
    if total_expense == 0: 
        return percent_dict

    for category in list(percent_dict["goal"].keys()): 
        this_category_expense = category_expense[category]
        # Calculate the percentage  
        percent_dict["actual"][category] = (this_category_expense / total_expense) * 100
        percent_dict["actual"][category] = round(percent_dict["actual"][category], 2)

    return percent_dict


def budget_progress_percentage(user: User, interval_type: str) -> Dict: 
    """ Calculate the the progress percentage of each towards that category's budget """

    # Budget plan and total budget
    plan = BudgetPlan.objects.get(user=user, interval_type=interval_type)
    total_budget = float(plan.recurring_income * plan.portion_for_expense / 100)
 
    # The total expense of each category dates of the current interval of given type
    first_date, last_date = get_curr_dates(interval_type) 
    category_expense = category_expense_dict(user, first_date, last_date)

    # Dictionary mapping each category to its current expense, budget, and its progress percentage 
    percent_dict = {
        "Expense": { "budget": total_budget }
    }

    # Add budget to the category first
    for category in list(plan.category_portion.keys()):
        this_category_budget = float(plan.category_portion[category])
        percent_dict[category] = { "budget": this_category_budget * total_budget / 100 }

    # Iterate through each key in the progress percentage dict to add current
    for category in list(percent_dict.keys()):
        # the current expense of this category 
        percent_dict[category]["current"] = category_expense[category]
        current_progress = percent_dict[category]["current"]
        budget_progress = percent_dict[category]["budget"]

        # calculate the progress percentage 
        if current_progress < budget_progress: 
            percentage = (current_progress / budget_progress) * 100
            percent_dict[category]["percentage"] = round(percentage, 2)
        else: 
            # otherwise, percentage is automatically 100%
            percent_dict[category]["percentage"] = 100

    return percent_dict


def get_budget_response_data(user: User, period_type: str) -> Dict: 
    """ 
    Get the custom response data used for budget plan, which includes the plan's 
    info along with its composition and progress percentage.
    """
    # Budget data
    budget_response = {}
    try: 
        budget_plan = BudgetPlan.objects.get(user=user, interval_type=period_type)
    except BudgetPlan.DoesNotExist: 
        # Return the empty dictionary if the queried plan doesn't exist
        return budget_response
    
    # composition percentage and progress percentage
    composition_dict = budget_composition_percentage(user, period_type)
    progress_dict = budget_progress_percentage(user, period_type)

    budget_response = {
        "id": budget_plan.id,   
        "income": budget_plan.recurring_income,
        "expense_portion": budget_plan.portion_for_expense,
        "composition": composition_dict, 
        "progress": progress_dict
    }
    
    return budget_response
