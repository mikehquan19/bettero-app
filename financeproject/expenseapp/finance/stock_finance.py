from decimal import Decimal
import decimal
from os import symlink
from typing import Dict
from datetime import date, timedelta
import yfinance as yf

# get the first date of last month and current date 
def get_first_and_last_dates(): 
    current_date = date.today() # the last date, which is today

    # the first date (which is first date of last month), month and year of the last date 
    prev_month, prev_year = current_date.month - 1, current_date.year
    if prev_month < 0: 
        prev_month, prev_year = 12, prev_year - 1

    first_date_last_month = date(year=prev_year, month=prev_month, day=1)
    return first_date_last_month, current_date


# convert the date obj to string 
def to_string(arg_date): 
    return f"{arg_date.year}-{arg_date.month}-{arg_date.day}"


# convert string to the date obj
def to_date(arg_str): 
    str_arr = arg_str.split("-")
    return date(year=int(str_arr[0]), month=int(str_arr[1]), day=int(str_arr[2]))


"""
Load the initial price of the stock since first date of last month
till the today along with current data of the stock 
"""
def load_stock_data(symbol: str) -> Dict: 
    # Get the first and last date 
    first_date, last_date = get_first_and_last_dates()

    # Load data of the stock's info 
    recent_data = yf.download(symbol, start=to_string(first_date), end=to_string(last_date))

    # current info of the stock 
    stock_data = {}
    for field in ["current_close", "previous_close", "open", "high", "low", "volume"]: 
        frame_col = field.split('_')[-1].capitalize()
        index = -2 if field == "previous_close" else -1
        value = recent_data[frame_col][symbol].iloc[index] 
        stock_data[field] = round(int(value), 2) if field == "volume" else round(Decimal(value), 2)

    # The price of the stock over the past
    stock_data["price_data"] = []
    current_date = first_date 
    while current_date < last_date: 
        try: 
            given_date_price = recent_data["Close"][symbol][to_string(current_date)]
            item_data = {
                "date": current_date, 
                "given_date_close": float(round(given_date_price, 2))
            }
            stock_data["price_data"].append(item_data)
        except KeyError: 
            # The key error means that the price of stock on that date doesn't exist
            pass
        
        current_date += timedelta(days=1) # Increment the date

    # The date the price of the stock was updated 
    last_updated_date = stock_data["price_data"][-1]["date"]
    stock_data["last_updated_date"] = last_updated_date

    return stock_data


# Update the info the stock, and add new record of the stock price 
def update_stock_data(symbol: str) -> Dict: 
    previous_date = to_string(date.today() - timedelta(days=1))
    today = to_string(date.today())

    # Return the Panda data frame 
    updated_data = yf.download(symbol, start=previous_date, end=today)

    custom_data = {}
    for field in ["new_close", "new_open", "new_high", "new_low", "new_volume"]: 
        frame_col = field.split('_')[-1].capitalize()
        value = updated_data[frame_col][symbol].iloc[0]
        value = int(value) if field == "new_volume" else Decimal(value)
        custom_data[field] = round(value, 2)

    return custom_data