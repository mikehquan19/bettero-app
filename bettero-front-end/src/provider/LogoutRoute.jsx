import { useEffect } from "react";
import { expenseappClient, REFRESH_TOKEN } from "./api";
import handleError from "./handleError";
import { Navigate } from "react-router-dom";


const LogoutRoute = () => {
    useEffect(() => {
        // blacklist of the refresh token so that it wont' be used
        const blacklistRefreshToken = async () => {
            const refreshToken = localStorage.getItem(REFRESH_TOKEN); 
            try {
                // call the API to blacklist the refresh token
                const response = await expenseappClient.post("/logout", {"refresh": refreshToken});
                console.log("log out request status: " + response.status); 

                // clear the local storage if there's nothing wroing.
                localStorage.clear();
                
            } catch (error) {
                handleError(error);
            }
        }
        blacklistRefreshToken(); 
    }, []);

    // navigate back to the login page
    return <Navigate to="/login" />
}

export default LogoutRoute;