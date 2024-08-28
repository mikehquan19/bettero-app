from datetime import timedelta, date
from .models import (
    Account,
    User, 
    Stock, 
    DateStockPrice, 
    Transaction, 
    OverdueBillMessage)
from . import finance 


# update the due date of the credit account 
def update_credit_due_date() -> None:
    try:
        # query the list of credit accounts 
        credit_account_list = Account.objects.filter(account_type="Credit")
        for account in credit_account_list: 
            if account.due_date >= date.today(): 
                account.due_date += timedelta(days=1)
                account.save()
    except Exception as e: 
        print("Error occur", e)


# update the info of the stock and create the record for the previous day
def update_info_and_create_price() -> None: 
    try: 
        # if it's sunday or monday, we won't need to check for the changed close
        # since the market is closed 
        weekday = date.today().weekday()
        if weekday != 0 and weekday != 6: 
            stock_list = Stock.objects.all()
            for stock in stock_list: 
                # fetch the updated info about the stock 
                updated_stock_data = finance.update_stock_data(stock.symbol)

                # update the data of each stock instance in the database
                # current close, previous_close, open, low, high, volume
                stock.previous_close = stock.current_close
                stock.current_close = updated_stock_data["new_close"]
                stock.open = updated_stock_data["new_open"]
                stock.low = updated_stock_data["new_low"]
                stock.high = updated_stock_data["new_high"]
                stock.volume = updated_stock_data["new_volume"]

                # update the last_updated_date 
                if stock.last_updated_date.weekday() == 4: 
                    stock.last_updated_date += timedelta(days=3)
                else: 
                    stock.last_updated_date += timedelta(days=1)
                stock.save() 

                # create the new date price instance for this stock 
                stock.datestockprice_set.create(
                    date=stock.last_updated_date, 
                    given_date_close=stock.current_close)
                
            # create the value of the portfolio 
            create_portfolio_value()
    except Exception as e: 
        print("Error occured", e)


# create the new portfolio value for the previous date 
def create_portfolio_value() -> None: 
    try: 
        # the date previous to the date in real time 
        previous_date = date.today() - timedelta(days=1)
        user_list = User.objects.all()

        for user in user_list: 
            # compute the total value of the user's portfolio 
            user_portfolio = user.stock_set.all()
            total_value = sum([(stock.current_close * stock.shares) for stock in user_portfolio])

            # create the portfolio value instance
            user.portfoliovalue_set.create(
                date=previous_date, 
                given_date_value=total_value)
    except Exception as e: 
        print("Error occured", e)


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
        print("Error occured", e)
        

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
        print("Error occured", e)


# delete overdue bills, and add the messages to the list 
def delete_overdue_bills_and_messages() -> None: 
    try:
        user_list = User.objects.all()
        for user in user_list: 
            # query the given user's list of overdue bills 
            overdue_bill_list = user.bills_set.filter(due_date__lt=date.today())

            # add the overdue message corresponding to the bills
            for overdue_bill in overdue_bill_list: 
                user.overduebillmessage_set.create(
                    bill_description=overdue_bill.description, 
                    bill_amount=overdue_bill.amount, 
                    bill_due_date=overdue_bill.due_date,
                    appear_date=date.today())
            # delete the queryset 
            overdue_bill_list.delete()
        
        # automatically delete the overdue bills message that are one day old
        # query the list of 1-day-old overdue messages and delete them 
        overdue_message_list = OverdueBillMessage.objects.filter(appear_date__lt=date.today())
        overdue_message_list.delete()
    except Exception as e: 
        print("Error occured", e)
