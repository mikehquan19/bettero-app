import { useForm } from "react-hook-form";
import { useEffect } from "react";
import { Link, useNavigate } from "react-router-dom";
import { expenseappClient, ACCESS_TOKEN, REFRESH_TOKEN } from "../../../provider/api";
import './LoginPage.css';

const LoginPage = ({type = "LOGIN"}) => {

    // function to log the error in the console depending on the type of error 
    const handleError = (error) => {
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

    // necessary components of useForm() hook 
    const { register, reset, handleSubmit, formState: { errors } } = useForm();
    const navigate = useNavigate(); 
    console.log(type);

    useEffect(() => {
        reset({}); 
    }, [])

    const handleSubmitButton = async (data, e) => {
        e.preventDefault(); 
        let submittedData; 
        if (type === "LOGIN") {
            submittedData = {
                username: data.username,
                password: data.password,
            };
        } else {
            submittedData = {...data};
        }
        console.log(submittedData);
        try {
            // call the API to get the access and refresh token
            const endpoint = type === "LOGIN" ? "/login/" : "/register/";
            const response = await expenseappClient.post(endpoint, submittedData); 
            localStorage.setItem(ACCESS_TOKEN, response.data.access); 
            localStorage.setItem(REFRESH_TOKEN, response.data.refresh); 
            if (type === "LOGIN") {
                navigate("/");
            } else {
                navigate("/login");
            }
        }
        catch (error) {
            handleError(error); 
        }
    }

    
    return (
        <>
            <div className="login-header-bar">Bettero App</div>
            <div className="login-form-wrapper" >
                <h3>{type === "LOGIN" ? <span>LOG IN</span> : <span>SIGN UP</span>}</h3>
                <form id="login" onSubmit={handleSubmit(handleSubmitButton)}>
                    {
                        type == "SIGNUP" && (
                            <>
                                <div className="form-field">
                                    <label htmlFor="full_name">Full name:</label>
                                    <small className="error-message">
                                        {errors?.full_name && errors?.full_name.message}
                                    </small>
                                    <input type="text" id="full_name" name="full_name"
                                        placeholder="Full name" {...register("full_name", {
                                            required: "Full name is required",
                                        })}/>
                                </div>
                                <div className="form-field">
                                    <label htmlFor="user_email">Email address:</label>
                                    <small className="error-message">
                                        {errors?.user_email && errors?.user_email.message}
                                    </small>
                                    <input type="email" id="user_email" name="user_email"
                                        placeholder="Email" {...register("user_email", {
                                            required: "Email is required", 
                                        })}/>
                                </div>
                            </>
                        )
                    }
                    <div className="form-field">
                        <label htmlFor="username">Username:</label>
                        <small className="error-message">
                            {errors?.username && errors?.username.message}
                        </small>
                        <input type="text" id="username" name="username"
                            placeholder="Username" {...register("username", {
                                required: "Username is required", 
                            })} />
                    </div>
                    <div className="form-field">
                        <label htmlFor="password">Password:</label>
                        <small className="error-message">
                            {errors?.password && errors?.password.message}
                        </small>
                        <input type="password" id="password" name="password" 
                            placeholder="Password" {...register("password", {
                                required: "Password is required", 
                            })}/>
                    </div>
                    {
                        type === "SIGNUP" && (
                            <div className="form-field">
                                <label htmlFor="password_again">Password again:</label>
                                <small className="error-message">
                                    {errors?.password_again && errors?.password_again.message}
                                </small>
                                <input type="password" id="password_again" name="password_again"
                                    placeholder="Password again" {...register("password_again", {
                                        required: "Password again is required",
                                    })}/>
                            </div>
                        )
                    }
                    {type === "LOGIN" ? (
                        <p>Don't have an account? <Link to="/register">Sign up</Link></p>
                    ) : (
                        <p>Already have an account? <Link to="/login">Log in</Link></p>
                    )}
                    <button type="submit">{type === "LOGIN" ? <span>Log in</span> : <span>Sign up</span>}</button>
                </form>
            </div>
        </>
    )
}

export default LoginPage;