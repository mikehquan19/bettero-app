import './../FormStyle.css';
import {useForm} from 'react-hook-form';
import { expenseappClient } from '../../../../provider/api';
import { faX } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';

const TransactionForm = ({ 
    accountListData, 
    handleChangeTransList, 
    handleHideForm}) => {
    // necessary components of useForm() hook 
    const { register, handleSubmit, formState: { errors } } = useForm();

    // submit the form's data 
    const handleSubmitButton = (data, e) => {
        e.preventDefault(); 
        // customize the submitted data 
        const chosenAccount = accountListData[data.account_index]; 
         
        // copy all properties of transaction form data except account-index
        const {account_index, ...copiedFormData} = data;
        const submittedData = {
            "account": chosenAccount.id,
            ...copiedFormData, 
        }
        submittedData.occur_date += new Date().toISOString().substring(10);
        // log in the console for checking 
        console.log(submittedData); 
        
        // call POST api to add transactions to the database
        expenseappClient.post("/transactions/", submittedData)
            .then((response) => {
                // for checking 
                console.log("add transaction request status: " + response.status); 
                handleChangeTransList(response.data); 
            })
            // log the error in the console depending on the type of error 
            .catch((error) => {
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
            }); 
        handleHideForm(); 
    }

    // validate some fields of the form 
    const registerOptions = {
        // validation on the description field 
        description: {
            required: "Description is required", 
            maxLength: {
                value: 200, 
                message: "Description must be less than 200 digits",
            }
        }, 
        // validation on the amount field
        amount: {
            required: "Amount is required",
            min: {
                value: 0.01, 
                message: "Amount must be at least 0.01",
            },
            maxLength: {
                value: 10, 
                message: "Amount must be less than 10 digits "
            }
        }
    }

    const handleSubmitError = (errors) => console.log(errors)

    // render the component 
    return (
        <>
            <div className="form-wrapper">
                <h3>ADD TRANSACTIONS</h3>
                <form id="add-transaction" onSubmit={handleSubmit(handleSubmitButton, handleSubmitError)}>
                    <div className="form-field">
                        <label htmlFor="account_index">Account name: </label>
                        <select name="account_index" {...register("account_index")}>
                            {accountListData.map((account, index) => 
                                <option key={index} value={index}>{account.name}</option>
                            )}
                        </select>
                    </div>
                    <div className="form-field">
                        <label htmlFor="description">Description: </label>
                        <small className="error-message">
                            {errors?.description && errors.description.message}
                        </small>
                        <input name="description" type="text" placeholder="Description"
                            {...register("description", registerOptions.description)} />
                    </div>
                    <div className="form-field">
                        <label htmlFor="amount">Amount: </label>
                        <small className="error-message">
                            {errors?.amount && errors.amount.message}
                        </small>
                        <input name="amount" type="text" placeholder="Amount"
                            {...register("amount", registerOptions.amount)} />
                    </div>
                    <div className="form-field">
                        <label htmlFor="from_account">Expense ? </label>
                        <input id="from-account" name="from_account" type="checkbox"
                            {...register("from_account")} />
                    </div>
                    <div className="form-field">
                        <label htmlFor="occur_date">Occur date: </label>
                        <input name="occur_date" type="date" {...register("occur_date")} />
                    </div>
                    <div className="form-field">
                        <label htmlFor="category">Category: </label>
                        <select name="category" {...register("category")}>
                            <option value="Grocery">Grocery</option>
                            <option value="Dining">Dining</option>
                            <option value="Shopping">Shopping</option>
                            <option value="Bills">Bills</option>
                            <option value="Gas">Gas</option>
                            <option value="Others">Others</option>
                        </select>
                    </div>
                </form>
                <div className="button-area">
                    <button type="submit" form="add-transaction" className="submit-button">Submit</button>
                    <button className="close-button" onClick={handleHideForm} >Close</button>
                </div>
                <button className="x-button" onClick={handleHideForm}><FontAwesomeIcon icon={faX} /></button>
            </div>
            <div className="overlay"></div>
        </>
    );
}

export default TransactionForm; 