import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { faX } from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { expenseappClient } from "../../../../provider/api";
import './../FormStyle.css';


// the form to add bill component 
const BillForm = ({
    type = "ADD",  
    currentFormData = null, 
    onChangeBillList = null, 
    onHideForm, 
}) => {

    // necessary components of useForm() hook 
    const { register, reset, handleSubmit, formState: { errors } } = useForm();
    const [payAccountList, setPayAccountList] = useState([]); 

    // function to load the form data if any
    const fetchFormData = () => {
        if (currentFormData !== null) {
            // pass the existing form data along with the userId
            reset({...currentFormData,});
        } 
    }

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

    // function to fetch the data about the account list
    const fetchPayAccountList = () => {  
        expenseappClient.get("/accounts/")
            .then((response) => {
                console.log("get account request status: " + response.status);
                setPayAccountList(response.data); 
            })
            .catch((error) => handleError(error)); 
    }
    // load all necessary data 
    useEffect(() => {
        fetchFormData();
        fetchPayAccountList(); 
    }, []); 

    const handleSubmitButton = (data, e) => {
        e.preventDefault(); 
        console.log("request data:");
        console.log(data); 
        // add new bill to the database 
        if (type === "ADD") {
            // call the POST method API to add the input bills 
            expenseappClient.post("/bills/", data)
                .then((response) => {
                    // for checking 
                    console.log("add bill request status: " + response.status); 
                    onChangeBillList(response.data); 
                })
                .catch((error) => console.log(error.message)); 
        } 
        // update the existing bill in the database 
        else if (type === "UPDATE") {
            // call the PUT method API to update the given bills 
            expenseappClient.put(`/bills/${currentFormData.id}/`, data)
                .then((response) => {
                    // info the updated bill for checking 
                    console.log("update bill request status: " + response.status); 

                    // call the GET method API to fetch new list of bills 
                    // and set it to the displayed list of bills 
                    expenseappClient.get(`/bills/`)
                        .then((response) => onChangeBillList(response.data))
                        .catch((error) => console.log(error.message)); 
                })
                .catch((error) => console.log(error.message)); 

        }
        onHideForm(); // hide the form 
    }

    // validation of each field in the form 
    // TODO: ADD the validation to the field of the form 
    const fieldValidation = {
        description: {
            required: "*Description is required", 
        },
        amount: {
            required: "*Amount is required", 
            min: {
                value: 0, 
                message: "*Balance must be greater than 0",
            },
        },
    }

    return (
        <>
            <div className="form-wrapper">
                <h3>{type} BILL</h3>
                <form id="add-bill" onSubmit={handleSubmit(handleSubmitButton)}>
                    <div className="form-field">
                        <label htmlFor="description">Description:</label>
                        <small className="error-message">
                            {errors?.description && errors.description.message}
                        </small>
                        <input type="text" placeholder="Description" name="description"
                        id="description" {...register("description", fieldValidation.description)}/>
                    </div>
                    <div className="form-field">
                        <label htmlFor="amount">Amount:</label>
                        <small className="error-message">
                            {errors?.amount && errors.amount.message}
                        </small>
                        <input type="number" step="any" placeholder="Amount" name="amount"
                        id="amount" {...register("amount", fieldValidation.amount)}/>
                    </div>
                    <div className="form-field">
                        <label htmlFor="pay_account">Pay account:</label>
                        <select name="pay_account" id="pay_account" {...register("pay_account")}>
                            {payAccountList.map(payAccount => 
                                <option key={payAccount.id} value={payAccount.id}>{payAccount.name}</option>
                            )}
                        </select>
                    </div>
                    <div className="form-field">
                        <label htmlFor="due_date">Due date</label>
                        <input type="date" id="due_date" 
                        name="due_date" {...register("due_date")}/>
                    </div>
                </form>
                <div className="button-area">
                    <button className="submit-button" form="add-bill">Submit</button>
                    <button className="close-button" onClick={onHideForm}>Close</button>
                </div>
                <button className="x-button" onClick={onHideForm}><FontAwesomeIcon icon={faX} /></button>
            </div>
            <div className="overlay"></div>
        </>
    );
}

export default BillForm;