# Generated by Django 5.1.4 on 2025-02-09 01:41

import expenseapp.models
from django.db import migrations, models


class Migration(migrations.Migration):

    dependencies = [
        ('expenseapp', '0003_remove_budgetplan_bills_remove_budgetplan_dining_and_more'),
    ]

    operations = [
        migrations.AlterField(
            model_name='budgetplan',
            name='category_portion',
            field=models.JSONField(default=expenseapp.models.get_default_dict),
        ),
    ]
