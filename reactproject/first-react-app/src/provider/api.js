import axios from 'axios'

export const ACCESS_TOKEN = "access";
export const REFRESH_TOKEN = "refresh";

// the instance of the axios used for this app
export const expenseappClient = axios.create({
    // "/choreo-apis/bettero-app/expense-app-service/v1/expenseapp/"
    baseURL: "http://127.0.0.1:8000/expenseapp/",
})

// interceptor to automatically add bearer token if there is any 
// before the request is sent 
expenseappClient.interceptors.request.use(
    (config) => {
        const token = localStorage.getItem(ACCESS_TOKEN); 
        if (token) {
            config.headers.Authorization = `Bearer ${token}`
        }
        return config; 
    }, 
    // if there is no tokens, throw an error
    (error) => {return Promise.reject(error);}
);
