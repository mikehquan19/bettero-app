import { useForm } from 'react-hook-form';
import { postUserTransactions } from '@provider/api';
import handleError from '@provider/handleError';
import { faX } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import './Form.scss'; 
import { Account } from '@interface';

interface TransactionFormProps {
  accountList: Account[], 
  onChangeTransList: (data: any) => void, 
  onHide: () => void
}

function TransactionForm({ accountList, onChangeTransList, onHide }: TransactionFormProps) {
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
      .then((response) => response ? onChangeTransList(response.data) : null)
      .catch((error) => handleError(error));
    
    onHide();
  }

  // validate some fields of the form 
  const registerOptions = {
    // validation on the description field 
    description: {
      required: "Description is required",
      maxLength: {
        value: 200,
        message: "Description must be less than 200 characters",
      }
    },
    // validation on the amount field
    amount: {
      required: "Amount is required",
      min: {
        value: 0.01,
        message: "Amount must be at least 0.01",
      },
      maxLength: {
        value: 10,
        message: "Amount must be less than 10 digits "
      }
    }
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
              {...register("description", registerOptions.description)} />
          </div>

          <div className="form-field">
            <label htmlFor="amount">Amount: </label>
            <small className="error-message">
              {errors?.amount && errors?.amount?.message as string}
            </small>
            <input type="text" placeholder="Amount"
              {...register("amount", registerOptions.amount)} />
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