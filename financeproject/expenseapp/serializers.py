import datetime
from . import models
from rest_framework import serializers
from django.core.exceptions import ValidationError
from . import finance

class RegisterSerializer(serializers.ModelSerializer): 
    class Meta: 
        model = models.User
        fields = ["full_name", "user_email", "username", "password", "password_again"]

    """
    {
        "full_name": "sample user", 
        "user_email": "sample.user@gmail.com", 
        "username": "sample username", 
        "password": "sample password", 
        "password_again": "sample password"
    }

    """

    # field of second password 
    # TODO: what is the write_only parameters ?
    password_again = serializers.CharField(max_length=20, required=True, write_only=True)

    # validate if 2 passwords the user entered match each other 
    def validate(self, attrs):
        if attrs.get('password') != attrs.get('password_again'): 
            raise serializers.ValidationError({"password": "2 passwords don't match."})
        return attrs

    # create the new user with the given validated info 
    def create(self, validated_data): 
        validated_data.pop("password_again")
        password = validated_data.pop("password")
        user = models.User.objects.create(**validated_data)

        # built-in method to hash the password of the user 
        user.set_password(password)
        user.save()
        return user


# the serializer of the account 
class AccountSerializer(serializers.ModelSerializer): 
    class Meta: 
        model = models.Account 
        fields = "__all__"
    """
        sample POST, PUT request data 
        {
            "user": 5,  
            "account_number": 6000,
            "name": "BOA's Debit Account",
            "institution": "BOA",
            "account_type": "Debit",
            "balance": "22.00"
            "credit_limit": null, 
            "due_date": null
        }
    """
    def update(self, instance, validated_data): 
        # compute the change of balance
        previous_balance = instance.balance
        current_balance = validated_data["balance"]
        balance_change = current_balance - previous_balance

        # create the transaction corresponding to the change, if any
        if balance_change != 0: 
            # The content of the transaction differs depending on change
            if balance_change > 0: 
                description = f"Account's balance increases ${abs(balance_change)}"
                from_account = True if instance.account_type == "Credit" else False
            else: 
                description = f"Account's balance decreases ${abs(balance_change)}"
                from_account = False if instance.account_type == "Credit" else True
            # create transaction
            models.Transaction.objects.create(
                user=instance.user,
                account=instance, 
                description=description, 
                amount=abs(balance_change),
                from_account=from_account,
                occur_date=datetime.datetime.now(),
                category="Others",
            )

        # call the original update() method 
        return super().update(instance, validated_data)

    # overidding the representation of the date field 
    def to_representation(self, instance):
        representation = super().to_representation(instance)
        if representation["due_date"]:
            representation["due_date"] = instance.due_date.strftime("%m/%d/%Y")
        return representation


# the serializer of the transaction 
class TransactionSerializer(serializers.ModelSerializer): 
    class Meta: 
        model = models.Transaction
        exclude = ["user"] 
    """ 
        sample POST request data
        {
            "account": 34,
            "description": "Starbucks' Orders",
            "amount": "22.50",
            "from_account": true,
            "occur_date": "2024-07-27",
            "category": "Dining"
        }
    """
    # create the account objects given the validated data
    def create(self, validated_data): 
        # create the transaction with the associated user, and account 
        account = validated_data.pop("account")
        user = account.user
        transaction = models.Transaction.objects.create(user=user, account=account, **validated_data)

        # adjust balance of the queried account 
        finance.adjust_account_balance(account, transaction)
        return transaction 


    # overidding the representation of the datetime field 
    def to_representation(self, instance):
        representation = super().to_representation(instance)
        representation["account_name"] = instance.account.name
        representation.pop("account")
        representation["occur_date"] = instance.occur_date.strftime("%m/%d/%Y")
        return representation


# the serializer of the budget plan
class BudgetPlanSerializer(serializers.ModelSerializer): 
    class Meta: 
        model = models.BudgetPlan
        fields = "__all__"
    """
        Sample POST, PUT data: 
        {
            "user": 5,
            "interval_type": "week", 
            "recurring_income": 4200, 
            "portion_for_expense": 75, 
            "grocery": 15.0,
            "dining": 15.0,
            "shopping": 10.0,
            "bills": 30.0,
            "gas": 10.0,
            "others": 10.0
        }
    """
    # update the account instance given the validated data from json 
    def update(self, instance, validated_data): 
        # only update if the instance and validated data has the same interval type
        if instance.interval_type != validated_data["interval_type"]: 
            raise ValidationError("The data has different interval type. Can't update.")
        
        # call the original update() method 
        return super().update(instance, validated_data)
    

# the serializer of the bills 
class BillSerializer(serializers.ModelSerializer): 
    class Meta: 
        model = models.Bills
        fields = "__all__"
    """
        Sample POST, PUT data
        {
            "user": 5, 
            "pay_account": 14, 
            "description": "Monthly Utility Bills City of Garland", 
            "amount": 230.00, 
            "due_date": "2024-08-16"
        }
    """
    # overidding the representation of the datetime field 
    def to_representation(self, instance):
        representation = super().to_representation(instance)
        representation["due_date"] = instance.due_date.strftime("%m/%d/%Y")
        representation["pay_account_name"] = instance.pay_account.name
        return representation
    

# the serializer of the overdue bill message 
class OverdueBillMessageSerializer(serializers.ModelSerializer): 
    class Meta: 
        model = models.OverdueBillMessage
        exclude = ["appear_date"] 

    def to_representation(self, instance):
        representation =  super().to_representation(instance)
        representation["bill_due_date"] = instance.bill_due_date.strftime("%m/%d/%Y")
        return representation
    

# the serializer of the price of stock on each date 
class StockPriceSerializer(serializers.ModelSerializer): 
    class Meta:
        model = models.DateStockPrice
        fields = "__all__"

    def to_representation(self, instance):
        representation =  super().to_representation(instance)
        representation["date"] = instance.date.strftime("%m/%d/%Y")
        return representation


# serializer of the stock 
class StockSerializer(serializers.ModelSerializer): 
    class Meta: 
        model = models.Stock
        fields = "__all__"

    """
        Sample POST, PUT DATA
        {
            "user": 5, 
            "corporation": "Apple.Inc",
            "name": "Apple stock",
            "symbol": "AAPL", 
            "shares": 20
        }
    """
    # restrict the field the user can update 
    def update(self, instance, validated_data): 
        # the user can't update 
        if validated_data["symbol"] != instance.symbol: 
            raise ValidationError("The data has different symbol. Can't update.")
        return super().update(instance, validated_data)

    # add change between 2 closes 
    def to_representation(self, instance):
        representation =  super().to_representation(instance)
        # change percentage from the previous close to the current close 
        representation.pop("previous_close")
        change = (instance.current_close - instance.previous_close) 
        representation["change"] = '{0:.2f}'.format(change)

        # change format of the date
        representation["last_updated_date"] = instance.last_updated_date.strftime("%m/%d/%Y")
        return representation
    

# serializer of the value of the porfolio
class PortfolioValueSerializer(serializers.ModelSerializer): 
    class Meta: 
        model = models.PortfolioValue
        fields = "__all__"

    def to_representation(self, instance):
        representation = super().to_representation(instance)
        representation["date"] = instance.date.strftime("%m/%d/%Y")
        return representation