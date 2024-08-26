import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { expenseappClient } from '../../../../provider/api';
import { faX } from '@fortawesome/free-solid-svg-icons';
import './../DeleteFormStyle.css';



// form to delete the plan 
const DeletePlanForm = ({intervalType, handleChangePlanList, handleHideForm}) => {

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
        if (e.target.value === "true") {
            expenseappClient.delete(`budget/${intervalType}`)
                .then((response) => {
                    console.log("delete plan request status: " + response.status)
                    // call the GET API to reset the list of budget plans 
                    expenseappClient.get("/budget/") 
                        .then((response) => handleChangePlanList(response.data))
                        .catch((error) => handleError(error)); 
                })
                .catch((planError) => handleError(planError)); 
        }
        handleHideForm(); 
    }

    return (
        <>
            <div className="delete-form-wrapper">
                <button className="x-button" onClick={handleHideForm} style={{
                    backgroundColor: "skyblue", 
                    color: "black",
                }}><FontAwesomeIcon icon={faX}/></button>
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

export default DeletePlanForm;