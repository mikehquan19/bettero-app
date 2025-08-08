import { ReactNode, useEffect } from "react";
import { expenseappClient, REFRESH_TOKEN } from "./api";
import handleError from "./handleError";
import { Navigate } from "react-router-dom";

/**
 * Logout route
 * @returns {ReactNode}
 */
export default function LogoutRoute(): ReactNode {
    /**
     * Blacklist of the refresh token so that it wont' be used
     * @returns {Promise<void>}
     */
    async function blacklistRefreshToken(): Promise<void> {
        const refreshToken = localStorage.getItem(REFRESH_TOKEN);
        try {
            const response = await expenseappClient.post("/logout", { "refresh": refreshToken });
            if (response.status === 200) {
                // clear the local storage if there's nothing wroing.
                localStorage.clear();
            }
        } catch (error) {
            handleError(error);
        }
    }

    useEffect(() => {
        blacklistRefreshToken(); 
    }, []);

    // navigate back to the login page
    return <Navigate to="/login" />
}