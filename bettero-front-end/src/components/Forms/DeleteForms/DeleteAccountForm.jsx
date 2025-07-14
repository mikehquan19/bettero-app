import { faX } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { expenseappClient } from '../../../provider/api.js';
import handleError from '../../../provider/handleError.js';
import './DeleteFormStyle.css';

// form to double-check whether user wants to delete account
function DeleteAccountForm ({accountId, accountType, handleChangeAccList, handleHideForm }) {

  function handleSubmitForm(e) {
    e.preventDefault();
    
    // if the user wants to delete account 
    if (e.target.value === "true") {
      expenseappClient.delete(`/accounts/${accountId}`)
        .then((response) => {
          if (response) {
            // call the get API to reset the list of accounts 
            expenseappClient.get("/accounts")
              .then((accountResponse) => {
                handleChangeAccList(accountResponse.data.filter(account => 
                  account.account_type === accountType
                ));
              })
              .catch((accountError) => handleError(accountError));
          }
        })
        .catch((error) => handleError(error));
    }
    handleHideForm();
  }

  return (
    <>
      <div className="delete-form-wrapper">
        <button className="x-button" onClick={handleHideForm}><FontAwesomeIcon icon={faX} /></button>
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