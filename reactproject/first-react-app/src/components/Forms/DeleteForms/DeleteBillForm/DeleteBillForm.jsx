import { faX } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { expenseappClient } from '../../../../provider/api';
import './../DeleteFormStyle.css';

// form to verify the decision to remove the bill
const DeleteBillForm = ({
    billId, 
    onChangeBillList, 
    onHideForm,
}) => {

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
        console.log("the id of the bill: " + billId); 
        if (e.target.value === "true") {
            // call the DELETE method API to delete the chosen bill
            expenseappClient.delete(`bills/${billId}`)
                .then((response) => {
                    console.log("delete bills request status: " + response.status);
                    // call the GET method API to get the new list of bills 
                    expenseappClient.get("/bills/")
                        .then((response) => {
                            // change to the new list of bills 
                            onChangeBillList(response.data); 
                        })
                        .catch(error => handleError(error));
                })
                .catch((billError) => handleError(billError)); 
        }
        onHideForm(); 
    }

    return (
        <>
            <div className="delete-form-wrapper">
            <button className="x-button" onClick={onHideForm}><FontAwesomeIcon icon={faX}/></button>
                <form>
                    <p style={{
                        marginTop: "5px",
                        fontSize: "14px", 
                        color: "red",
                    }}>The payment will be automatically added to list of transactions.</p>
                    <label style={{fontSize: "18px"}}>Do you want to delete this bill ?</label>
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

export default DeleteBillForm;