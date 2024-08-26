import { expenseappClient } from "../../../../provider/api";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faX } from '@fortawesome/free-solid-svg-icons';
import './../DeleteFormStyle.css';

const DeleteStockForm = ({ 
    stockSymbol, 
    onChangeStockList, 
    onHideForm}) => {

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
        console.log(stockSymbol);
        if (e.target.value === "true") {
            expenseappClient.delete(`/stocks/${stockSymbol}/`)
                .then((response) => {
                    console.log("stock to be deleted " + response.data.message);
                    // call the GET method API to fetch new list of stocks 
                    expenseappClient.get(`/stocks/`)
                        .then((response) => onChangeStockList(response.data))
                        .catch((error) => handleError(error)); 
                })
                .catch((stockError) => handleError(stockError)); 
        }
        onHideForm(); 
    }

    return (
        <>
            <div className="delete-form-wrapper">
                <button className="x-button" onClick={onHideForm} style={{
                        backgroundColor: "skyblue", 
                        color: "black",
                    }}><FontAwesomeIcon icon={faX}/></button>
                <form style={{
                    marginTop: "1em",
                }}>
                    <label>Do you want to delete this stock ?</label>
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

export default DeleteStockForm;