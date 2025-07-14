import { faX } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { expenseappClient } from '../../../provider/api';
import handleError from '../../../provider/handleError';
import './DeleteFormStyle.css';

// form to verify the decision to remove the bill
function DeleteBillForm ({ billId, onChangeBillList, onHideForm }) {
  function handleSubmitForm(e) {
    e.preventDefault();
    
    if (e.target.value === "true") {
      // call the DELETE method API to delete the chosen bill
      expenseappClient.delete(`bills/${billId}`)
        .then((response) => {
          if (response) {
            expenseappClient.get("/bills")
              .then((response) => onChangeBillList(response.data))
              .catch(error => handleError(error));
          }
        })
        .catch((billError) => handleError(billError));
    }
    onHideForm();
  }

  return (
    <>
      <div className="delete-form-wrapper">
        <button className="x-button" onClick={onHideForm}><FontAwesomeIcon icon={faX} /></button>
        <form>
          <p style={{
            marginTop: "5px",
            fontSize: "14px",
            color: "red",
          }}>The payment will be automatically added to list of transactions.</p>
          <label style={{ fontSize: "18px" }}>Do you want to delete this bill ?</label>
          
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