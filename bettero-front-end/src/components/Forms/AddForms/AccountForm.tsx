import { useEffect } from 'react';
import handleError from '@provider/handleError';
import { useForm } from 'react-hook-form';
import { faX } from '@fortawesome/free-solid-svg-icons';
import { postOrPutUserAccount } from '@provider/api';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { ADD_ACCOUNT_VALIDATION } from '@Forms/AddForms/Validation';
import { Account } from '@interface';
import './FormStyle.css';

interface AccountFormProps {
  type: "ADD" | "UPDATE", 
  isCreditAccount: boolean, 
  currentData?: Account, 
  onChangeAccInfo?: (data: Account) => void,
  onChangeAccList?: (data: Account[]) => void, 
  onHide: () => void
}

// the form to add component
function AccountForm({
  type = "ADD", isCreditAccount, currentData, onChangeAccInfo, onChangeAccList, onHide 
}: AccountFormProps) {

  const { register, reset, handleSubmit, formState: { errors } } = useForm();

  // Set initial value for the debit account form 
  useEffect(() => {
    if (!!currentData) {
      // pass the current data as the default values of the form 
      reset({ ...currentData, dueDate: currentData.dueDate?.toString() });
    } else {
      if (!isCreditAccount) {
        reset((previousData) => ({...previousData, accountType: "Debit"}));
      } else {
        reset((previousData) => ({...previousData, accountType: "Credit"}));
      }
    }

  }, []);

  // handle the submission of the form 
  function handleSubmitButton(data: any) {
    console.log(data);
    if (type == 'ADD') { 
      postOrPutUserAccount('ADD', data)
        .then((response) => {
          if (!!onChangeAccList) {
            const accountType = isCreditAccount ? 'Credit' : 'Debit'; 
            const newAccountList = response.data as Account[]; 
            onChangeAccList(newAccountList.filter(account => account.accountType === accountType));
          }
        })
        .catch((error) => handleError(error));
    } 
    else if (type === 'UPDATE') { 
      postOrPutUserAccount('UPDATE', data, currentData!.id)
        .then((response) => !!onChangeAccInfo ? onChangeAccInfo(response.data): null)
        .catch((error) => handleError(error));
    }
    onHide();
  }

  return (
    <>
      <div className="form-wrapper">
        <h3>{type} {isCreditAccount ? 'CREDIT' : 'DEBIT'} ACCOUNTS</h3>
        <form style={{ marginTop: "20px" }} id="add-account" onSubmit={handleSubmit(handleSubmitButton)}>
          <div className="form-field">
            <label htmlFor="account-number">Account number: </label>
            <small className="error-message">
              {errors?.accountNumber?.message && errors?.accountNumber?.message as string}
            </small>
            <input type="number" placeholder='Account number'
              {...register("accountNumber", ADD_ACCOUNT_VALIDATION.accountNumber)} />
          </div>

          <div className="form-field">
            <label htmlFor="account-name">Account name: </label>
            <small className="error-message">
              {errors?.name?.message && errors?.name?.message as string}
            </small>
            <input type="text" placeholder='Account name'
              {...register("name", ADD_ACCOUNT_VALIDATION.name)} />
          </div>

          <div className="form-field">
            <label htmlFor="institution">Institution: </label>
            <small className="error-message">
              {errors?.institution && errors?.institution?.message as string}
            </small>
            <input type="text" placeholder='Institution'
              {...register("institution", ADD_ACCOUNT_VALIDATION.institution)} />
          </div>

          <div className="form-field">
            <label htmlFor="balance">Balance: </label>
            <small className="error-message">
              {errors?.balance && errors?.balance?.message as string}
            </small>
            <input type="number" step="0.01" placeholder='Balance'
              {...register("balance", ADD_ACCOUNT_VALIDATION.balance)} />
          </div>

          { /* if it's credit account, then we have 2 more fields */
            (isCreditAccount) && (
              <>
                <div className="form-field">
                  <label htmlFor="credit-limit">Credit limit: </label>
                  <small className="error-message">
                    {errors?.creditLimit && errors?.creditLimit?.message as string}
                  </small>
                  <input type="number" step="0.01" placeholder='Credit limit'
                    {...register("creditLimit", ADD_ACCOUNT_VALIDATION.creditLimit)} />
                </div>

                <div className="form-field">
                  <label htmlFor="due-date">Due date: </label>
                  <input type="date" {...register("dueDate", ADD_ACCOUNT_VALIDATION.dueDate)} />
                </div>
              </>
            )
          }
        </form>

        <div className="button-area">
          <button type="submit" form="add-account" className="submit-button">Submit</button>
          <button className="close-button" onClick={onHide}>Close</button>
        </div>
        <button className="x-button" onClick={onHide}><FontAwesomeIcon icon={faX} /></button>
      </div>
      <div className="overlay"></div>
    </>
  );
}

export default AccountForm;