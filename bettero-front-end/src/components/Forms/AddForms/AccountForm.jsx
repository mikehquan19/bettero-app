import './FormStyle.css';
import { useEffect } from 'react';
import handleError from '../../../provider/handleError';
import { useForm } from 'react-hook-form';
import { faX } from '@fortawesome/free-solid-svg-icons';
import { expenseappClient } from '../../../provider/api';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { ADD_ACCOUNT_VALIDATION } from './Validation';

// the form to add component
export default function AccountForm({
  type = "ADD", isCreditAccount, currentData = null,
  handleChangeAccInfo = null, handleChangeAccList = null, handleHideForm 
}) {

  // necessary components of useForm() hook 
  const { register, reset, handleSubmit, formState: { errors } } = useForm();

  // set some initial value for the debit account form 
  useEffect(() => {
    // pass the current data as the default values of the form 
    if (!currentData) {
      reset({ ...currentData, });
    } else {
      if (!isCreditAccount) {
        reset((previousData) => ({...previousData, account_type: "Debit"}))
      } else {
        reset((previousData) => ({...previousData, account_type: "Credit"}))
      }
    }
  }, []);

  // handle the submission of the form 
  function handleSubmitButton(data, e) {
    e.preventDefault();

    // if the form is to add accounts, call the POST API 
    if (type == "ADD") {
      expenseappClient.post("/accounts", data)
        .then((response) => handleChangeAccList(response.data))
        .catch((error) => handleError(error));
    }
    // otherwise, call the PUT API 
    else if (type === "UPDATE") {
      expenseappClient.put(`/accounts/${currentData.id}`, data)
        .then((response) => handleChangeAccInfo(response.data))
        .catch((error) => handleError(error));
    }
    handleHideForm();
  }

  return (
    <>
      <div className="form-wrapper">
        <h3>{type} {isCreditAccount ? (<span>CREDIT</span>) : (<span>DEBIT</span>)} ACCOUNTS</h3>
        <form style={{ marginTop: "20px" }} id="add-account" onSubmit={handleSubmit(handleSubmitButton)}>
          <div className="form-field">
            <label htmlFor="account-number">Account number: </label>
            <small className="error-message">
              {errors?.account_number && errors.account_number.message}
            </small>
            <input type="number" placeholder='Account number' name="account_number"
              {...register("account_number", ADD_ACCOUNT_VALIDATION.account_number)} />
          </div>

          <div className="form-field">
            <label htmlFor="account-name">Account name: </label>
            <small className="error-message">
              {errors?.name && errors.name.message}
            </small>
            <input type="text" placeholder='Account name' name="name"
              {...register("name", ADD_ACCOUNT_VALIDATION.name)} />
          </div>

          <div className="form-field">
            <label htmlFor="institution">Institution: </label>
            <small className="error-message">
              {errors?.institution && errors.institution.message}
            </small>
            <input type="text" placeholder='Institution' name="institution"
              {...register("institution", ADD_ACCOUNT_VALIDATION.institution)} />
          </div>

          <div className="form-field">
            <label htmlFor="balance">Balance: </label>
            <small className="error-message">
              {errors?.balance && errors.balance.message}
            </small>
            <input type="number" step="0.01" placeholder='Balance' name="balance"
              {...register("balance", ADD_ACCOUNT_VALIDATION.balance)} />
          </div>

          { /* if it's credit account, then we have 2 more fields */
            (isCreditAccount) && (
              <>
                <div className="form-field">
                  <label htmlFor="credit-limit">Credit limit: </label>
                  <small className="error-message">
                    {errors?.credit_limit && errors.credit_limit.message}
                  </small>
                  <input type="number" step="0.01" placeholder='Credit limit' name="credit_limit"
                    {...register("credit_limit", ADD_ACCOUNT_VALIDATION.credit_limit)} />
                </div>

                <div className="form-field">
                  <label htmlFor="due-date">Due date: </label>
                  <input type="date" name="due_date" {...register("due_date", register.due_date)} />
                </div>
              </>
            )
          }
        </form>

        <div className="button-area">
          <button type="submit" form="add-account" className="submit-button">Submit</button>
          <button className="close-button" onClick={handleHideForm}>Close</button>
        </div>
        <button className="x-button" onClick={handleHideForm}><FontAwesomeIcon icon={faX} /></button>
      </div>
      <div className="overlay"></div>
    </>
  );
}