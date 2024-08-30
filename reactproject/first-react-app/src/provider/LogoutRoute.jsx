import { useEffect } from "react";
import { ACCESS_TOKEN, expenseappClient, REFRESH_TOKEN } from "./api";
import { Navigate } from "react-router-dom";


const LogoutRoute = () => {

    useEffect(() => {
        // blacklist of the refresh token so that it wont' be used
        // clear the tokens from the local storage
        const blacklistRefreshToken = async () => {
            const refreshToken = localStorage.getItem(REFRESH_TOKEN); 
            const data = {"refresh": refreshToken};
            try {
                // call the API to blacklist the refresh token
                const response = await expenseappClient.post("/logout/", data);
                console.log("log out request status: " + response.status); 
                
            } catch (error) {
                if (error.response) {
                    // The request was made and the server responded with a status code
                    // that falls out of the range of 2xx
                    console.log(error.response.data);
                    console.log(error.response.status);
                    console.log(error.response.headers);
                } else if (error.request) {
                    // The request was made but no response was received
                    console.log(error.request);
                } else {
                    // Something happened in setting up the request that triggered an Error
                    console.log('Error', error.message);
                }
                console.log(error.config); 
            }
            // clear the local storage 
            localStorage.clear(); 
            // test the ACCESS and REFRESH TOKEN
            console.log(localStorage.getItem(ACCESS_TOKEN)); 
            console.log(localStorage.getItem(REFRESH_TOKEN));
        }
        blacklistRefreshToken(); 
    }, []);

    // navigate back to the login page
    return <Navigate to="/login" />
}

export default LogoutRoute;