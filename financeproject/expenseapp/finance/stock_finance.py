from decimal import Decimal
from os import symlink
from typing import Dict
from datetime import date, timedelta
import yfinance as yf
from .utils import get_first_and_last_dates

def to_string(date: date) -> str: 
    """ Convert a date object to a string """

    month = f"0{date.month}" if date.month < 10 else str(date.month)
    day = f"0{date.day}" if date.day < 10 else str(date.day)
    return f"{date.year}-{month}-{day}"


def to_date(str: str) -> date: 
    """ Convert a string to a date object """
    splitted_str = str.split("-")
    return date(
        year=int(splitted_str[0]), 
        month=int(splitted_str[1]), 
        day=int(splitted_str[2])
    )


def load_stock_data(symbol: str) -> Dict: 
    """
    Load the initial price of the stock since first date of last month until the today 
    along with current data of the stock 
    """

    # Get the first and last date of a stock price's duration 
    first_date, last_date = get_first_and_last_dates()
    # Load data of the stock's info, returns Panda dataframe 
    recent_data = yf.download([symbol], start=to_string(first_date), end=to_string(last_date))
    # current info of the stock 
    stock_data = {}
    for field in ["current_close", "previous_close", "open", "high", "low", "volume"]: 
        dataframe_col = field.split("_")[-1].capitalize()
        index = -2 if field == "previous_close" else -1
        value = recent_data[dataframe_col][symbol].iloc[index]
        if field == "volume":  
            # Volume of the stock is defined as long
            stock_data[field] = int(value)
        else: 
            stock_data[field] = round(Decimal(value), 2)

    # The price of the stock over the past
    stock_data["price_data"] = []
    current_date = first_date 
    while current_date < last_date: 
        try: 
            given_date_price = recent_data["Close"][symbol][to_string(current_date)]
            stock_data["price_data"].append({
                "date": current_date, 
                "given_date_close": float(round(given_date_price, 2))
            })
        except KeyError: 
            # The key error means that the price of stock on that date doesn't exist
            pass
        current_date += timedelta(days=1) # Increment the date

    # The date the price of the stock was updated 
    last_updated_date = stock_data["price_data"][-1]["date"]
    stock_data["last_updated_date"] = last_updated_date

    return stock_data


def update_stock_data(symbol: str) -> Dict: 
    """ Update the info the stock, and add new record of the stock price """

    prev_date = to_string(date.today() - timedelta(days=1))
    today = to_string(date.today())
    # Return the Panda data frame 
    updated_data = yf.download(symbol, start=prev_date, end=today)

    custom_data = {}
    for field in ["new_close", "new_open", "new_high", "new_low", "new_volume"]: 
        frame_col = field.split("_")[-1].capitalize()
        value = updated_data[frame_col][symbol].iloc[0]
        if field == "volume": 
            custom_data[field] = int(value)
        else: 
            custom_data[field] = round(Decimal(value), 2)

    return custom_data