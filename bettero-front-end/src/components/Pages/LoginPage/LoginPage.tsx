import { useForm } from "react-hook-form";
import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { ACCESS_TOKEN, REFRESH_TOKEN, loginUser, registerUser } from "@provider/api";
import handleError from "@provider/handleError";
import './LoginPage.css';

interface LoginPageProps {
  type: string
}

function LoginPage({ type = "LOGIN" }: LoginPageProps) {
  // necessary components of useForm() hook 
  const { register, reset, handleSubmit, formState: { errors } } = useForm();
  const [isInvalidCredentials, setIsInvalidCredentials] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    reset({});
  }, [])

  async function handleSubmitButton(data: any) {
    if (type === "LOGIN") {
      try {
        // call the API to get the access and refresh token
        const response = await loginUser(data.username, data.password);
        if (response.status === 200) {
          localStorage.setItem(ACCESS_TOKEN, response.data.access);
          localStorage.setItem(REFRESH_TOKEN, response.data.refresh);
          navigate("/");
        }
      }
      catch (error: any) {
        handleError(error);
        if (error.response.data.detail === "No active account found with the given credentials") {
          setIsInvalidCredentials(true);
        }
      }
    } else {
      try {
        const response = await registerUser(data);
        if (response.status === 200) {
          navigate("/login");
        }
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
        <h3>{type}</h3>
        {(isInvalidCredentials && type == "LOGIN") &&
          <div style={{ textAlign: "center" }}>
            <small className="error-message">Invalid username or password!</small>
          </div>
        }
        <form id="login" onSubmit={handleSubmit(handleSubmitButton)}>
          {
            type == "SIGNUP" && (
              <>
                <div className="form-field">
                  <label htmlFor="full_name">Full name:</label>
                  <small className="error-message">
                    {errors?.fullName && errors?.fullName?.message as string}
                  </small>
                  <input type="text" id="full_name" placeholder="Full name" {...register("fullName", {
                      required: "Full name is required",
                    })} />
                </div>
                <div className="form-field">
                  <label htmlFor="user_email">Email address:</label>
                  <small className="error-message">
                    {errors?.userEmail && errors?.userEmail?.message as string}
                  </small>
                  <input type="email" id="user_email" placeholder="Email" {...register("userEmail", {
                      required: "Email is required",
                    })} />
                </div>
              </>
            )
          }
          <div className="form-field">
            <label htmlFor="username">Username:</label>
            <small className="error-message">
              {errors?.username && errors?.username?.message as string}
            </small>
            <input type="text" id="username" placeholder="Username" {...register("username", {
                required: "*Username is required",
              })} />
          </div>
          <div className="form-field">
            <label htmlFor="password">Password:</label>
            <small className="error-message">
              {errors?.password && errors?.password?.message as string}
            </small>
            <input type="password" id="password" placeholder="Password" {...register("password", {
                required: "*Password is required",
              })} />
          </div>
          {
            type === "SIGNUP" && (
              <div className="form-field">
                <label htmlFor="password_again">Password again:</label>
                <small className="error-message">
                  {errors?.passwordAgain && errors?.passwordAgain?.message as string}
                </small>
                <input type="password" id="password_again" 
                  placeholder="Password again" {...register("passwordAgain", {
                    required: "*Password again is required",
                  })} />
              </div>
            )
          }
          {
            type === "LOGIN" ? 
              <p>Don't have an account? <Link to="/register">Sign up</Link></p> : 
              <p>Already have an account? <Link to="/login">Log in</Link></p>
          }
          <button type="submit">{type === "LOGIN" ? "Log in" : "Sign up"}</button>
        </form>
      </div>
    </>
  )
}

export default LoginPage;