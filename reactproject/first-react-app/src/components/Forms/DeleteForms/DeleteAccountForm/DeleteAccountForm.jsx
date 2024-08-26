import { faX } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { expenseappClient } from '../../../../provider/api.js';
import './../DeleteFormStyle.css';

// form to double-check whether user wants to delete account
const DeleteAccountForm = ({ 
    accountId, 
    accountType, 
    handleChangeAccList, 
    handleHideForm }) => {

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

    const handleSubmitForm = (e) => {
        e.preventDefault();
        console.log(e.target.value);  
        // if the user wants to delete account 
        if (e.target.value === "true") {
            // URL for the delete API 
            expenseappClient.delete(`/accounts/${accountId}/`)
                .then((response) => {
                    console.log(response.data.message); 
                    // call the get API to reset the list of accounts 
                    expenseappClient.get("/accounts/")
                        .then((accountResponse) => {
                            handleChangeAccList(accountResponse.data.filter(account => 
                                account.account_type === accountType
                            ));
                        })
                        .catch((accountError) => handleError(accountError));
                })
                .catch((error) => handleError(error));
        }
        handleHideForm();
    }

    return (
        <>
            <div className="delete-form-wrapper">
            <button className="x-button" onClick={handleHideForm}><FontAwesomeIcon icon={faX}/></button>
                <form>
                    <label>Do you want to delete the account ?</label>
                    <div className="delete-form-input">
                        <button onClick={handleSubmitForm} type="submit" value="true">YES</button>
                        <button onClick={handleSubmitForm} type="submit" value="false">NO</button>
                    </div>
                </form>
            </div>
            <div className="overlay"></div>
        </>
    );
}

export default DeleteAccountForm;