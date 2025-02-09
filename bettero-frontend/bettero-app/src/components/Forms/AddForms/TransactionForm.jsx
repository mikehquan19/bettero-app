import { useForm } from 'react-hook-form';
import { expenseappClient } from '../../../provider/api';
import handleError from '../../../provider/handleError';
import { faX } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import './FormStyle.css';

const TransactionForm = ({
  accountListData,
  handleChangeTransList,
  handleHideForm 
}) => {
  // necessary components of useForm() hook 
  const { register, handleSubmit, formState: { errors } } = useForm();

  // submit the form's data 
  const handleSubmitButton = (data, e) => {
    e.preventDefault();
    // customize the submitted data 
    const account = accountListData[data.account_index];

    // copy all properties of transaction form data except account-index
    const { account_index, ...copiedFormData } = data;
    const submittedData = {"account": account.id, ...copiedFormData}
    submittedData.occur_date += new Date().toISOString().substring(10);

    // call POST api to add transactions to the database
    expenseappClient.post("/transactions", submittedData)
      .then((response) => {
        handleChangeTransList(response.data);
      })
      // log the error in the console depending on the type of error 
      .catch((error) => handleError(error));
    handleHideForm();
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
            <select name="account_index" {...register("account_index")}>
              {accountListData.map((account, index) =>
                <option key={index} value={index}>{account.name}</option>
              )}
            </select>
          </div>

          <div className="form-field">
            <label htmlFor="description">Description: </label>
            <small className="error-message">
              {errors?.description && errors.description.message}
            </small>
            <input name="description" type="text" placeholder="Description"
              {...register("description", registerOptions.description)} />
          </div>

          <div className="form-field">
            <label htmlFor="amount">Amount: </label>
            <small className="error-message">
              {errors?.amount && errors.amount.message}
            </small>
            <input name="amount" type="text" placeholder="Amount"
              {...register("amount", registerOptions.amount)} />
          </div>

          <div className="form-field">
            <label htmlFor="occur_date">Occur date: </label>
            <input name="occur_date" type="date" {...register("occur_date")} />
          </div>

          <div className="form-field">
            <label htmlFor="category">Category: </label>
            <select name="category" {...register("category")}>
              {categoryArr.map(category => <option id={category} value={category}>{category}</option>)}
            </select>
          </div>
        </form>

        <div className="button-area">
          <button type="submit" form="add-transaction" className="submit-button">Submit</button>
          <button className="close-button" onClick={handleHideForm} >Close</button>
        </div>
        <button className="x-button" onClick={handleHideForm}><FontAwesomeIcon icon={faX} /></button>
      </div>
      <div className="overlay"></div>
    </>
  );
}

export default TransactionForm; 