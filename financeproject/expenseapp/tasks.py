from django.db import transaction
from datetime import timedelta, date
from .models import (
    Account, PortfolioValue, User, Stock, DateStockPrice, 
    Transaction, OverdueBillMessage)
from .finance import update_stock_data
import traceback

# update the due date of the credit account 
def update_credit_due_date() -> None:
    try:
        # query the list of credit accounts 
        credit_account_list = Account.objects.filter(account_type="Credit", due_date__lte=date.today())
        updated_account_list = []
        with transaction.atomic(): 
            for i, credit_account in enumerate(credit_account_list): 
                updated_account_list.append(credit_account)
                # increment the month of the due date by 1 (same day next month)
                updated_account_list[i].due_date.month += 1 
            # bulk-update() will make stuff more efficient with 1 query 
            if len(updated_account_list) > 0: 
                num_updated_account = Account.objects.bulk_update(updated_account_list, ["due_date"])
                print(f"{num_updated_account} accounts updated successfully")
            else: 
                print("No accounts updated!")
    except Exception: 
        # print the traceback of the errors 
        traceback.print_exc()


# update the info of the stock and create the record for the previous day
def update_info_and_create_price() -> None: 
    try: 
        # if it's sunday or monday, we won't need to check for the changed close
        # since the market is closed 
        weekday = date.today().weekday()
        if weekday != 0 and weekday != 6: 
            with transaction.atomic(): 
                stock_list = Stock.objects.all()
                updated_stock_list = []
                updated_field_list = ["previous_close", "current_close", "open", "low", "high", "volume", "last_updated_date"]

                for i, stock in enumerate(stock_list): 
                    # fetch the updated info about the stock 
                    updated_stock_data = update_stock_data(stock.symbol)
                    updated_stock_list.append(stock)

                    # update the data of each stock instance in the database
                    # current close, previous_close, open, low, high, volume
                    updated_stock_list[i].previous_close = stock.current_close
                    updated_stock_list[i].current_close = updated_stock_data["new_close"]
                    updated_stock_list[i].open = updated_stock_data["new_open"]
                    updated_stock_list[i].low = updated_stock_data["new_low"]
                    updated_stock_list[i].high = updated_stock_data["new_high"]
                    updated_stock_list[i].volume = updated_stock_data["new_volume"]

                    # update the last_updated_date 
                    if updated_stock_list[i].last_updated_date.weekday() == 4: 
                        updated_stock_list[i].last_updated_date += timedelta(days=3)
                    else: 
                        updated_stock_list[i].last_updated_date += timedelta(days=1)
                
                # using bulk_update() to update with only 1 query
                if len(updated_stock_list) > 0: 
                    num_updated_stock = Stock.objects.bulk_update(updated_stock_list, updated_field_list)
                    updated_stock_queryset = Stock.objects.filter(id__in=[stock.id for stock in updated_stock_list])
                    print(f"{num_updated_stock} stocks updated successfully!")
                else: 
                    updated_stock_queryset = Stock.objects.none()
                    print("No stocks updated!")

                # create the new date price instance for the updated stock
                created_stock_price_list = []
                for updated_stock in updated_stock_queryset: 
                    created_stock_price_list.append(DateStockPrice(
                        stock=updated_stock, 
                        date=updated_stock.last_updated_date, 
                        given_date_close=updated_stock.current_close,
                    ))
                # bulk_create() to make it more efficient 
                if len(created_stock_price_list) > 0: 
                    created_queryset = DateStockPrice.objects.bulk_create(created_stock_price_list)
                    print(f"{len(created_queryset)} dates stock price created")
                else: 
                    print("Nothing created.")

                # create the value of the portfolio 
                create_portfolio_value()
    except Exception as e: 
        traceback.print_exc()


# create the new portfolio value for the previous date 
def create_portfolio_value() -> None: 
    try: 
        # the date previous to the date in real time 
        previous_date = date.today() - timedelta(days=1)
        user_list = User.objects.all()
        created_portfolio_value_list = []

        with transaction.atomic(): 
            for user in user_list: 
                # compute the total value of the user's portfolio 
                user_portfolio = Stock.objects.filter(user=user)
                total_value = sum([(stock.current_close * stock.shares) for stock in user_portfolio])
                created_portfolio_value_list.append(PortfolioValue(
                    user=user, date=previous_date, given_date_value=total_value
                ))
            if len(created_portfolio_value_list) > 0: 
                PortfolioValue.objects.bulk_create(created_portfolio_value_list)
    except Exception as e: 
        traceback.print_exc()


# delete the list of date prices that were beyond first day of last month
def delete_price() -> None: 
    try:
        # compute the first date of last month
        current_month = date.today().month 
        current_year = date.today().year 
        prev_month = current_month - 1
        prev_year = current_year
        if prev_month < 0: 
            prev_month = 12
            prev_year -= 1
        first_date_last_month = date(year=prev_year, month=prev_month, day=1)

        # query the list of price instance and delete 
        old_date_price_list = DateStockPrice.objects.filter(date__lt=first_date_last_month)
        old_date_price_list.delete()
    except Exception as e: 
        traceback.print_exc()
        

# delete the list of transactions that 
def delete_transactions() -> None: 
    try:
        # compute the first date of 5 months ago 
        first_date_this_month = date(year=date.today().year, month=date.today().month, day=1)
        filter_date = first_date_this_month - timedelta(days=155)

        # query the list of transactions that are 5months old and delete
        old_transaction_list = Transaction.objects.filter(occur_date__lt=filter_date)
        old_transaction_list.delete()
    except Exception as e: 
        traceback.print_exc()


# delete overdue bills, and add the messages to the list 
def delete_overdue_bills_and_messages() -> None: 
    try:
        user_list = User.objects.all()
        with transaction.atomic():
            for user in user_list: 
                # query the given user's list of overdue bills 
                overdue_bill_list = user.bills_set.filter(due_date__lt=date.today())
                created_overdue_message_list = []

                # add the overdue message corresponding to the bills
                for overdue_bill in overdue_bill_list: 
                    created_overdue_message_list.append(OverdueBillMessage(
                        user=user, 
                        bill_description=overdue_bill.description, 
                        bill_amount=overdue_bill.amount, 
                        bill_due_date=overdue_bill.due_date,
                        appear_date=date.today()
                    ))
                if len(created_overdue_message_list) > 0: 
                    OverdueBillMessage.objects.bulk_create(created_overdue_message_list)
                # delete the queryset 
                overdue_bill_list.delete()
        
        # automatically delete the overdue bills message that are one day old
        # query the list of 1-day-old overdue messages and delete them 
        overdue_message_list = OverdueBillMessage.objects.filter(appear_date__lt=date.today())
        overdue_message_list.delete()
    except Exception as e: 
        traceback.print_exc()
