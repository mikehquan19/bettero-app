import { faX } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { expenseappClient, getUserAccounts } from '@provider/api';
import handleError from '@provider/handleError';
import { Account } from '@interface';
import './DeleteFormStyle.css';

interface DeleteAccountFormInterface {
  accountId: number, 
  accountType: string, 
  onChangeAccList: (data: Account[]) => void, 
  onHide: () => void
}

// form to double-check whether user wants to delete account
function DeleteAccountForm ({
  accountId, accountType, onChangeAccList, onHide
}: DeleteAccountFormInterface) {

  function handleSubmitForm(e: any) {
    e.preventDefault();
    
    if (e.target.value === "true") {
      // if the user wants to delete account 
      expenseappClient.delete(`/accounts/${accountId}`)
        .then((response1) => {
          if (response1) {
            getUserAccounts()
              .then((response2) => {
                onChangeAccList(response2.data.filter((account: Account) => 
                  account.accountType === accountType));
              })
              .catch((accountError) => handleError(accountError));
          }
        })
        .catch((error) => handleError(error));
    }
    onHide();
  }

  return (
    <>
      <div className="delete-form-wrapper">
        <button className="x-button" onClick={onHide}>
          <FontAwesomeIcon icon={faX} />
        </button>
        <form>
          <label>Do you want to delete the account ?</label>
          <div className="delete-form-input">
            <button onClick={handleSubmitForm} type="submit" value="true">
              YES</button>
            <button onClick={handleSubmitForm} type="submit" value="false">
              NO</button>
          </div>
        </form>
      </div>
      <div className="overlay"></div>
    </>
  );
}

export default DeleteAccountForm;