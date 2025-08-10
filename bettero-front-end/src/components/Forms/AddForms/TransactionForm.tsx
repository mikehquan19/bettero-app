import { useForm } from 'react-hook-form';
import { postUserTransactions } from '@provider/api';
import handleError from '@provider/handleError';
import { faX } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import './Form.scss'; 
import { Account } from '@interface';
import { ADD_TRANSACTION_VALIDATION } from './Validation';

interface TransactionFormProps {
  accountList: Account[], 
  onChangeTransList: (data: any) => void, 
  onHide: () => void
}

/**
 * Form to add or update the transactions 
 * @param {TransactionFormProps} props 
 * @returns {JSX.Element}
 */
function TransactionForm(
  { accountList, onChangeTransList, onHide }: TransactionFormProps
): JSX.Element {
  const { register, handleSubmit, formState: { errors } } = useForm();

  // submit the form's data 
  function handleSubmitButton(data: any): void {
    // Customize the submitted data 
    // Copy all properties of transaction form data except account index
    const { accountIndex, ...copiedFormData } = data;
    const submittedData = {
      account: accountList[data.accountIndex as number].id, // account Id
      ...copiedFormData
    };
    submittedData.occurDate += new Date().toISOString().substring(10);

    postUserTransactions(submittedData)
      .then((response) => {
        if (response) onChangeTransList(response.data);
      })
      .catch((error) => handleError(error));
    
    onHide();
  }

  const categoryArr = ["Housing", "Automobile", "Medical", "Subscription", "Grocery", "Dining", "Shopping", "Gas", "Others", "Income"];

  // render the component 
  return (
    <>
      <div className="form-wrapper">
        <h3>ADD TRANSACTIONS</h3>
        <form id="add-transaction" onSubmit={handleSubmit(handleSubmitButton, (errors) => handleError(errors))}>

          <div className="form-field">
            <label htmlFor="account_index">Account name: </label>
            <select {...register("accountIndex")}>
              {accountList.map((account, index) =>
                <option key={index} value={index}>{account.name}</option>
              )}
            </select>
          </div>

          <div className="form-field">
            <label htmlFor="description">Description: </label>
            <small className="error-message">
              {errors?.description && errors?.description?.message as string}
            </small>
            <input type="text" placeholder="Description"
              {...register("description", ADD_TRANSACTION_VALIDATION.description)} />
          </div>

          <div className="form-field">
            <label htmlFor="amount">Amount: </label>
            <small className="error-message">
              {errors?.amount && errors?.amount?.message as string}
            </small>
            <input type="text" placeholder="Amount"
              {...register("amount", ADD_TRANSACTION_VALIDATION.amount)} />
          </div>

          <div className="form-field">
            <label htmlFor="occur_date">Occur date: </label>
            <input type="date" {...register("occurDate")} />
          </div>

          <div className="form-field">
            <label htmlFor="category">Category: </label>
            <select {...register("category")}>
              {categoryArr.map(category => 
                <option key={category} id={category} value={category}>{category}</option>)
              }
            </select>
          </div>
        </form>

        <div className="button-area">
          <button type="submit" form="add-transaction" className="submit-button">Submit</button>
          <button className="close-button" onClick={onHide} >Close</button>
        </div>
        <button className="x-button" onClick={onHide}><FontAwesomeIcon icon={faX} /></button>
      </div>
      <div className="overlay"></div>
    </>
  );
}

export default TransactionForm; 