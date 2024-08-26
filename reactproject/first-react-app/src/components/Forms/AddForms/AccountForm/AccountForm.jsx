import './../FormStyle.css';
import { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { faX } from '@fortawesome/free-solid-svg-icons';
import { expenseappClient } from '../../../../provider/api';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';

// the form to add component
const AccountForm = ({ 
    type = "ADD",
    isCreditAccount, 
    currentData = null, 
    handleChangeAccInfo = null,
    handleChangeAccList = null, 
    handleHideForm }) => {

    // necessary components of useForm() hook 
    const { register, reset, handleSubmit, formState: { errors } } = useForm();

    // set some initial value for the debit account form 
    useEffect(() => {
        // pass the current data as the default values of the form 
        if (currentData !== null) {
            reset({...currentData,}); 
        } else {
            if (!isCreditAccount) {
                reset((previousData) => ({
                    ...previousData,
                    account_type: "Debit",
                }))
            } else {
                reset((previousData) => ({
                    ...previousData, 
                    account_type: "Credit",
                }))
            }
        }
    }, []);

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

    // handle the submission of the form 
    const handleSubmitButton = (data, e) => {
        e.preventDefault();
        // for checking, log the submitted data in the console
        console.log(data);
        
        // if the form is to add accounts, call the POST API 
        if (type == "ADD") {
            expenseappClient.post("/accounts/", data)
                .then((response) => {
                    // reset the the account list 
                    console.log("add account request status: " + response.status);
                    handleChangeAccList(response.data);
                })
                // log the error in the console depending on the type of error 
                .catch((error) => handleError(error));
        }
        // otherwise, call the PUT API 
        else if (type === "UPDATE") {
            // checking if the API endpoint is correct or not 
            expenseappClient.put(`/accounts/${currentData.id}/`, data)
                .then((response) => {
                    // reset the data the updated account 
                    console.log("reset account request status: " + response.status);
                    handleChangeAccInfo(response.data);
                })
                .catch((error) => handleError(error));
        }
        // hide the form 
        handleHideForm();
    }

    // the validation on each field of the form 
    const registerOptions = {
        account_number: { 
            required: "*Account number is required", 
            minLength: {
                value: 2, 
                message: "*Account number must have at least 2 digits",
            },
            maxLength: {
                value: 12,
                message: "*Account number only has at most 12 digits", 
            }
        },
        name: { required: "*Account name is required"},
        institution: { required: "*Institution is required"},
        balance: {
            required: "*Balance is required",
            min: {
                value: 0, 
                message: "*Balance must be greater than 0",
            }
        },
        credit_limit: {
            required: "*Credit limit is required", 
            min: {
                value: 0, 
                message: "*Credit limit must be greater than 0",
            }
        },
    }

    return (
        <>
            <div className="form-wrapper">
                <h3>{type} {isCreditAccount ? (<span>CREDIT</span>) : (<span>DEBIT</span>)} ACCOUNTS</h3>
                <form style={{ marginTop: "20px", }} id="add-account"
                    onSubmit={handleSubmit(handleSubmitButton)}>
                    <div className="form-field" >
                        <label htmlFor="account-number">Account number: </label>
                        <small className="error-message">
                            {errors?.account_number && errors.account_number.message}
                        </small>
                        <input type="number" placeholder='Account number' name="account_number"
                            {...register("account_number", registerOptions.account_number)} />
                    </div>
                    <div className="form-field">
                        <label htmlFor="account-name">Account name: </label>
                        <small className="error-message">
                            {errors?.name && errors.name.message}
                        </small>
                        <input type="text" placeholder='Account name' name="name"
                            {...register("name", registerOptions.name)} />
                    </div>
                    <div className="form-field">
                        <label htmlFor="institution">Institution: </label>
                        <small className="error-message">
                            {errors?.institution && errors.institution.message}
                        </small>
                        <input type="text" placeholder='Institution' name="institution"
                            {...register("institution", registerOptions.institution)} />
                    </div>
                    <div className="form-field">
                        <label htmlFor="balance">Balance: </label>
                        <small className="error-message">
                            {errors?.balance && errors.balance.message}
                        </small>
                        <input type="number" step="0.01" placeholder='Balance' name="balance"
                            {...register("balance", registerOptions.balance)} />
                    </div>
                    {/* if it's credit account, then we have 2 more fields */}
                    {
                        (isCreditAccount) && (
                            <>
                                <div className="form-field">
                                    <label htmlFor="credit-limit">Credit limit: </label>
                                    <small className="error-message">
                                        {errors?.credit_limit && errors.credit_limit.message}
                                    </small>
                                    <input type="number" step="0.01" placeholder='Credit limit' name="credit_limit"
                                        {...register("credit_limit", registerOptions.credit_limit)} />
                                </div>
                                <div className="form-field">
                                    <label htmlFor="due-date">Due date: </label>
                                    <input type="date" name="due_date"
                                        {...register("due_date", register.due_date)} />
                                </div>
                            </>
                        )
                    }
                </form>
                <div className="button-area">
                    <button type="submit" form="add-account" className="submit-button">Submit</button>
                    <button className="close-button" onClick={handleHideForm}>Close</button>
                </div>
                <button className="x-button" onClick={handleHideForm}><FontAwesomeIcon icon={faX}/></button>
            </div>
            <div className="overlay"></div>
        </>
    );
}

export default AccountForm;