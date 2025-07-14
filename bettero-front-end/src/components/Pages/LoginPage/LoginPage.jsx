import { useForm } from "react-hook-form";
import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { expenseappClient, ACCESS_TOKEN, REFRESH_TOKEN } from "../../../provider/api";
import handleError from "../../../provider/handleError";
import './LoginPage.css';

const LoginPage = ({ type = "LOGIN" }) => {
  // necessary components of useForm() hook 
  const { register, reset, handleSubmit, formState: { errors } } = useForm();
  const [isInvalidCredentials, setIsInvalidCredentials] = useState(false);
  const navigate = useNavigate();
  console.log(type);

  useEffect(() => {
    reset({});
  }, [])

  const handleSubmitButton = async (data, e) => {
    e.preventDefault();
    let submittedData;
    if (type === "LOGIN") {
      try {
        // call the API to get the access and refresh token
        const response = await expenseappClient.post("/login", {
          username: data.username, 
          password: data.password, 
        });
        console.log("login request status: " + response.status);
        localStorage.setItem(ACCESS_TOKEN, response.data.access);
        localStorage.setItem(REFRESH_TOKEN, response.data.refresh);
        navigate("/");
      }
      catch (error) {
        handleError(error);
        // notify the user 
        if (error.response.data.detail === "No active account found with the given credentials") {
          setIsInvalidCredentials(true);
        }
      }
    } else {
      try {
        // call the API to create new user 
        const response = await expenseappClient.post("/register", data);
        console.log("create new user request status: " + response.status);
        navigate("/login");
      }
      catch (error) { 
        handleError(error) 
      }
    }
  }


  return (
    <>
      <div className="login-header-bar">Bettero App</div>
      <div className="login-form-wrapper" >
        <h3>{type === "LOGIN" ? <span>LOG IN</span> : <span>SIGN UP</span>}</h3>
        {(isInvalidCredentials && type == "LOGIN") &&
          <div style={{ textAlign: "center" }}><small className="error-message">Invalid username or password!</small></div>
        }
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
                    })} />
                </div>
                <div className="form-field">
                  <label htmlFor="user_email">Email address:</label>
                  <small className="error-message">
                    {errors?.user_email && errors?.user_email.message}
                  </small>
                  <input type="email" id="user_email" name="user_email"
                    placeholder="Email" {...register("user_email", {
                      required: "Email is required",
                    })} />
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
                required: "*Username is required",
              })} />
          </div>
          <div className="form-field">
            <label htmlFor="password">Password:</label>
            <small className="error-message">
              {errors?.password && errors?.password.message}
            </small>
            <input type="password" id="password" name="password"
              placeholder="Password" {...register("password", {
                required: "*Password is required",
              })} />
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
                    required: "*Password again is required",
                  })} />
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