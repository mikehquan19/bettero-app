# Bettero App

The full-stack CRUD app that allows users to
- keep track of their financial accounts, transactions, bills, stocks, etc. 

- create their own budget plan and see how their expense behavior has gone so far compared to the plan. 

- visualize their expense behavior (and also their stock portfolios) through different types of charts.

## Tech Stack
- [Django (Django Rest Framework)](https://www.django-rest-framework.org/): Backend

- [Celery & Celery Beat](https://docs.celeryq.dev/en/stable/django/index.html): For concurrent and periodic execution of the tasks since the app will need to periodically update credit due date, stock prices, and more.

- [React](https://react.dev/): Frontend

- [ChartJS](https://www.chartjs.org/): For the visualization. 

The app also implements [JWT Authentication](https://towardsdev.com/a-comprehensive-guide-to-jwt-for-identity-authentication-and-authorization-in-asp-net-703f5af195d1) for login/logout system

## Improvement in the future 
Some improvements that I'm actively working on: 
- Integration with [Plaid API](https://plaid.com/docs/api/), which would make the app hugely automatic and is a good plus.

- Better design in general.

## How to run

Front-end is available [here](https://github.com/mikehquan19/bettero-frontend). Please clone them along with this repo.

The app will be deployed soon, and there is a demo available in [youtube](https://www.youtube.com/watch?v=J6bK8G2iXxw). But below is the guide to run them locally: 

### 1/ Prerequisites
Before getting started, make sure that you have [Node.JS](https://nodejs.org/en) & [Django](https://www.djangoproject.com/) installed. If not, please visit the websites for the download and installation guide. 

Or run the following commands: 
```
pip install django
npm install node
```

Make sure you also have a `PostgreSQL` server up and running, and `RabbitMQ` broker for Celery service to run on. 


### 2/ Set up and run the project

Clone the project 
```
https://github.com/mikehquan19/bettero-app.git
```

#### For backend, 
To be added 

#### For frontend,

Finally, run the project 
```
cd bettero-frontend
cd bettero-app
npm run dev 
```

## Contact
If you have any questions regarding this project, please reach out to me at hoangphucquan19@gmail.com