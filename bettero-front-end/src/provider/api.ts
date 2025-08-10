import { Transaction } from '@interface';
import axios, {AxiosResponse} from 'axios';
import humps from 'humps';

const IS_DEPLOYED = false;
const LOCALHOST_URL = 'http://localhost:8000/expenseapp';
const DEPLOYMENT_URL = '/choreo-apis/bettero-app/expense-app-service/v1/expenseapp/';
export const ACCESS_TOKEN = "access";
export const REFRESH_TOKEN = "refresh";

/**
 * Instance of the axios used for this app
 */
export const expenseappClient = axios.create({
    baseURL: IS_DEPLOYED ? DEPLOYMENT_URL : LOCALHOST_URL
});

// interceptor to automatically add bearer token if there is any, before the request is sent 
expenseappClient.interceptors.request.use(
    (config) => {
        const token = localStorage.getItem(ACCESS_TOKEN); 
        if (token) {
            config.headers.Authorization = `Bearer ${token}`
        }
        return config; 
    }, 
    // if there is no tokens, throw an error
    (error) => { return Promise.reject(error); }
);

/** 
 * Get financial summary of user 
 */
export async function getFinancialSummary() {
    return expenseappClient.get('/summary')
        .then((response) => humps.camelizeKeys(response) as AxiosResponse<any, any>);
}

/** 
 * Get the list of accounts of user 
 */
export async function getUserAccounts() {
    return expenseappClient.get('/accounts')
        .then((response) => humps.camelizeKeys(response) as AxiosResponse<any, any>);; 
}

/**
 * Post or put user account
 */
export async function postOrPutUserAccount(type: 'ADD' | 'UPDATE', data: any, id?: number) {
    const promise = type === 'ADD' ? 
        expenseappClient.post('/accounts', humps.decamelizeKeys(data)) : 
        expenseappClient.put(`/accounts/${id}`, humps.decamelizeKeys(data)); 
    
    return promise.then(response => humps.camelizeKeys(response) as AxiosResponse<any, any>);
}

/** 
 * Get the list of paginagted transactions of user 
 */
export async function getUserTransactions(pageIndex?: number) {
    const endpoint = !pageIndex || pageIndex === 1 ? '/transactions' : `/transactions?page=${pageIndex}`;
    return expenseappClient.get(endpoint)
        .then((response) => humps.camelizeKeys(response) as AxiosResponse<any, any>)
        .then((response) => ({
            status: response.status, 
            data: {
                "page": pageIndex ? pageIndex : 1,
                "transactionCount": response.data.count as number, 
                "transactionList": response.data.results as Transaction[]
            }
        }));
}

/**
 * Post new transaction to the list
 */
export async function postUserTransactions(data: any) {
    return expenseappClient.post("/transactions", humps.decamelizeKeys(data))
        .then((response) => humps.camelizeKeys(response) as AxiosResponse<any, any>)
        .then((response) => ({
            status: response.status, 
            data: {
                page: 1,
                transactionCount: response.data.count, 
                transactionList: response.data.results as Transaction[]
            }
        }));
}

/** 
 * Get full financial summary of user 
 */
export async function getFullFinancialSummary() {
    return expenseappClient.get('/full_summary')
        .then((response) => humps.camelizeKeys(response) as AxiosResponse<any, any>); 
}

/** 
 * Get the list of bills of user 
 */
export async function getBills() {
    return expenseappClient.get('/bills')
        .then((response) => humps.camelizeKeys(response) as AxiosResponse<any, any>); 
}

export async function postOrPutBill(type: 'ADD' | 'UPDATE', data: any) {
    const promise = type === 'ADD' ? 
        expenseappClient.post("/bills", humps.decamelizeKeys(data)) : 
        expenseappClient.put(`/bills/${data.id}`, humps.decamelizeKeys(data));
    
    return promise.then((response) => humps.camelizeKeys(response) as AxiosResponse<any, any>);
}

/** 
 * Get the budget data of user 
 */
export async function getBudgetPlan() {
    return expenseappClient.get('/budget')
        .then((response) => humps.camelizeKeys(response) as AxiosResponse<any, any>);
}

/** 
 * Get the list of overdue messages of user 
 */
export async function getOverdueMessages() {
    return expenseappClient.get('/overdue_message')
        .then((response) => humps.camelizeKeys(response) as AxiosResponse<any, any>); ;
}

/** 
 * Get the list of stocks of user 
 */
export async function getStocks() {
    return expenseappClient.get('/stocks')
        .then((response) => humps.camelizeKeys(response) as AxiosResponse<any, any>);
}

/**
 * Add the stock to new list of stock 
 * @param {'ADD' | 'UPDATE'} type
 * @param {any} data
 */
export async function postOrPutStock(type: 'ADD' | 'UPDATE', data: any) {
    const promise = type === 'ADD' ?
        expenseappClient.post('/stocks', humps.decamelizeKeys(data)) : 
        expenseappClient.put(`/stocks/${data.symbol}`, humps.decamelizeKeys(data));
    
    return promise.then((response) => humps.camelizeKeys(response) as AxiosResponse<any, any>);
}

/** 
 * Get the value of user's porfolio starting from start of last month 
 */
export async function getPorfolioValues() {
    return expenseappClient.get('/portfolio_value')
        .then((response) => humps.camelizeKeys(response) as AxiosResponse<any, any>); ;
}

/**
 * Get the list of transactions of given category.
 * @param {string} selectedCategory 
 */
export async function getCategoryTransactions(selectedCategory: string, pageIndex?: number) {
    const endpoint = !pageIndex || pageIndex === 1 ? 
        `/transactions/category/${selectedCategory}` : 
        `/transactions/category/${selectedCategory}?page=${pageIndex}`;
    
    return expenseappClient.get(endpoint)
        .then((response) => humps.camelizeKeys(response) as AxiosResponse<any, any>)
        .then((response) => ({
            status: response.status, 
            data: {
                "page": pageIndex ? pageIndex : 1,
                "transactionCount": response.data.count, 
                "transactionList": response.data.results as Transaction[]
            }
        }));
}

/**
 * Get the financial summary of the account with given id.
 * @param {number} id: account's id
 */
export async function getAccountSummary(id: number) {
    return expenseappClient.get(`/accounts/${id}/summary`)
        .then((response) => humps.camelizeKeys(response) as AxiosResponse<any, any>);
}

/**
 * Get either the list of latest transactions or transactions of given category
 * from a account with given id.
 * 
 * @param {number} id: account's id
 * @param {string} transactionType: type of transaction, latest or category
 * @param {string} category: category of transactions
 */
export async function getAccountTransactions(
    id: number, transactionType: string="latest", category?: string, pageIndex?: number
) {
    let endpoint: string;
    if (transactionType === "latest") {
        endpoint = !pageIndex || pageIndex === 1 ? 
            `accounts/${id}/transactions` : `accounts/${id}/transactions?page=${pageIndex}`;
    } else if (transactionType === "category") {
        endpoint = !pageIndex || pageIndex === 1 ? 
            `accounts/${id}/transactions/both?category=${category}` : 
            `accounts/${id}/transactions/both?category=${category}&page=${pageIndex}`;
    } else {
        throw Error("Invalid transaction's type");
    }

    return expenseappClient.get(endpoint)
        .then((response) => humps.camelizeKeys(response) as AxiosResponse<any, any>)
        .then((response) => ({
            status: response.status, 
            data: {
                page: pageIndex ? pageIndex : 1,
                transactionCount: response.data.count, 
                transactionList: response.data.results as Transaction[]
            }
        }));
}

/**
 * Get the list of transactions of given category during a given interval 
 * @param {string} type: type of list of transactions: "interval" or "both"
 * @param {string} category 
 * @param {string} firstDate: first date of the interval 
 * @param {string} lastDate: last date of the interval
 */
export async function getSummaryTransactions(
    type: string, category?: string, firstDate?: string, lastDate?: string, pageIndex?: number
) {
    let endpoint; 
    if (type === "interval") {
      endpoint = `/transactions/interval?first_date=${firstDate}&last_date=${lastDate}`;
    } else if (type === "both") {
      endpoint = 
        `/transactions/both?category=${category}&first_date=${firstDate}&last_date=${lastDate}`;
    } else {
      throw new Error("Transaction type invalid!");
    }
    if (pageIndex && pageIndex !== 1) endpoint += `&page=${pageIndex}`;

    return expenseappClient.get(endpoint)
        .then((response) => humps.camelizeKeys(response) as AxiosResponse<any, any>)
        .then((response) => ({
            status: response.status, 
            data: {
                page: pageIndex ? pageIndex : 1,
                transactionCount: response.data.count, 
                transactionList: response.data.results as Transaction[]
            }
        }));
  }

/**
 * Get the list of price of stock from the beginning of last month.
 * @param {string} symbol
 */
export async function getStockPrice(symbol: string) {
    return expenseappClient.get(`/stocks/${symbol}`)
        .then((response) => humps.camelizeKeys(response) as AxiosResponse<any, any>);
}

/**
 * Login the user with the given username and password 
 * @param {string} username 
 * @param {string} password 
 * @returns 
 */
export function loginUser(username: string, password: string) {
    return expenseappClient.post('/login', {username: username, password: password});
}

export function registerUser(data: any) {
    return expenseappClient.post('/register', humps.decamelizeKeys(data));
}