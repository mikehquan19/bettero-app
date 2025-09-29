from rest_framework import status
from rest_framework.views import APIView
from rest_framework.permissions import IsAuthenticated
from rest_framework.response import Response
from django.shortcuts import get_object_or_404
from django.db import transaction
from expenseapp.models import DateStockPrice, Stock, PortfolioValue
from expenseapp.serializers import PortfolioValueSerializer, StockSerializer, StockPriceSerializer
from expenseapp.finance import load_stock_data

class StockList(APIView): 
    """ View to handle the list of stocks  """
    permission_classes = [IsAuthenticated]

    def get_response_data(self, request):
        """ Get the response data  """
        
        stock_list = Stock.objects.filter(user=request.user)
        response_data = StockSerializer(stock_list, many=True).data
        return response_data
    
    def get(self, request, format=None) -> Response: 
        """ GET method, return the list of stock of the user """
        response_data = self.get_response_data(request)
        return Response(response_data)
     
    def post(self, request, format=None) -> Response: 
        """ POST method, add the stock the list of stock of user """
        request_data = { **request.data, "user": request.user.id }

        # load the price data of the stock with the given symbol 
        try: 
            stock_data = load_stock_data(request_data["symbol"])
        except IndexError: 
            # This indicates that the stock with the given symbol isn't found
            return Response({"error": "No stock with the given symbol"}, status=status.HTTP_404_NOT_FOUND)
        
        stock_price_data = stock_data.pop("price_data")
        for data_field in list(stock_data.keys()): 
            request_data[data_field] = stock_data[data_field]

        # add to the list 
        new_stock_serializer = StockSerializer(data=request_data)
        if new_stock_serializer.is_valid(): 
            created_stock = new_stock_serializer.save() 

            # add all of the price of the stock
            stock_price_list = [] 
            for i in range(len(stock_price_data)): 
                stock_price_list.append(DateStockPrice(
                    stock=created_stock, 
                    date=stock_price_data[i]["date"], 
                    given_date_close=stock_price_data[i]["given_date_close"]
                ))
            with transaction.atomic(): 
                DateStockPrice.objects.bulk_create(stock_price_list)

            response_data = self.get_response_data(request)
            return Response(response_data, status=status.HTTP_201_CREATED)
        
        return Response(new_stock_serializer.errors, status=status.HTTP_400_BAD_REQUEST)
    

class StockPriceDetail(APIView): 
    """ View to handle the price detail of the stock  """

    permission_classes = [IsAuthenticated]

    def get_response_data(self, request, symbol): 
        stock = get_object_or_404(Stock, user=request.user, symbol=symbol)
        # List of prices of the stock 
        stock_price_list = stock.datestockprice_set.order_by("date")

        # custom reponse data for stock detail 
        response_data = {
            "stock": StockSerializer(stock).data, 
            "price_list": {}
        }
        price_list = StockPriceSerializer(stock_price_list, many=True).data
        for price in price_list: 
            response_data["price_list"][price["date"]] = price["given_date_close"]
    
        return response_data
    
    def get(self, request, symbol, format=None) -> Response: 
        """ GET method, return the detail of the stock, including its list of price """
        response_data = self.get_response_data(request, symbol)
        return Response(response_data)
    
    def put(self, request, symbol, format=None) -> Response: 
        """ PUT method, update the stock """
        request_data = { **request.data, "user": request.user.id }

        stock = get_object_or_404(Stock, user=request.user, symbol=symbol)
        updated_stock_serializer = StockSerializer(stock, data=request_data)
        
        if updated_stock_serializer.is_valid(): 
            updated_stock_serializer.save() 

            response_data = self.get_response_data(request, symbol)
            return Response(response_data, status=status.HTTP_202_ACCEPTED)
        
        return Response(updated_stock_serializer.errors, status=status.HTTP_400_BAD_REQUEST)
    
    def delete(self, request, symbol, format=None) -> Response: 
        """ DELETE method, delete the stock """

        stock = get_object_or_404(Stock, user=request.user, symbol=symbol)
        stock.delete()
        return Response({ "message": "Stock deleted successfully" }, status=status.HTTP_204_NO_CONTENT)
    

class PortfolioValueList(APIView): 
    permission_classes = [IsAuthenticated] 

    def get(self, request, format=None) -> Response: 
        portfolio_value_list = PortfolioValue.objects.filter(user=request.user).order_by("date")
        original_data = PortfolioValueSerializer(portfolio_value_list, many=True).data 

        response_data = {}
        for data_item in original_data: 
            response_data[data_item["date"]] = data_item["given_date_value"]
        
        return Response(response_data)