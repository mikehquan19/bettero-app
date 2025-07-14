import { useState, useEffect } from 'react';
import { jwtDecode } from 'jwt-decode';
import {REFRESH_TOKEN, ACCESS_TOKEN, expenseappClient } from './api';
import handleError from './handleError';
import { Navigate } from 'react-router-dom';

function PrivateRoute({children}) {
    // determine whether the user is logged in 
    const [isAuthorized, setIsAuthorized] = useState(null);

    // refresh the token if the access token expired
    async function refreshToken() {
        // the refresh token to be submitted alongside request
        const refreshToken = localStorage.getItem(REFRESH_TOKEN); 
        try {
            const response = await expenseappClient.post("/login/refresh", {refresh: refreshToken}); 
            if (response.status === 200) {
                // if request successful, reset the access token, user is authorized
                localStorage.setItem(ACCESS_TOKEN, response.data.access);
                setIsAuthorized(true); 
            }
            else { 
                setIsAuthorized(false); 
            }
        } 
        catch (error) { 
            // if there is error, user not authorized
            handleError(error);
            setIsAuthorized(false); 
        }
    }

    // process the access token to see if user is authorized
    async function processAuth() {
        const accessToken = localStorage.getItem(ACCESS_TOKEN);
        if (!accessToken) { // if access token nonexistent, user not authorized 
            setIsAuthorized(false);
        }
        else { // otherwise, decode the token and check if it expired
            const decodedAccessToken = jwtDecode(accessToken);
            const tokenExpirationTime = decodedAccessToken.exp;

            if (tokenExpirationTime < Date.now() / 1000) {
                // if token expires, try refreshing the token
                await refreshToken();
            }
            else {
                // otherwise, user is authorized 
                setIsAuthorized(true);
            }
        }
    }

    // call processAuth to see if the user is authorized 
    useEffect(() => processAuth(), []);

    // conditional rendering depending on if the user is authorized 
    if (isAuthorized === null) return <div>Loading...</div>

    // if the user is authorized, routed to main pages (chilren)
    // otherwise, redirect back to the login page
    return isAuthorized ? children : <Navigate to="/login" />
}

export default PrivateRoute; 