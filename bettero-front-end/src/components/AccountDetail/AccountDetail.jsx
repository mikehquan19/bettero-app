import TransactionTable from '../TransactionTable/TransactionTable';
import DeleteAccountForm from '../Forms/DeleteForms/DeleteAccountForm';
import AccountForm from '../Forms/AddForms/AccountForm';
import ChangeChart from '../Charts/ChangeChart';
import CompositionChart from '../Charts/CompositionChart';
import { getElementAtEvent } from 'react-chartjs-2';
import { reformatDate } from '../../utils';
import './AccountDetail.css';

export default function AccountDetail({ accountInfo, type = "Debit", handleChangeAccList }) {
  // info of the account 
  const [accountData, setAccountData] = useState({ ...accountInfo });

  // the expense change and composition data for the charts 
  const [expenseChange, setExpenseChange] = useState({});
  const [expenseComposition, setExpenseComposition] = useState({});

  // state of the transaction table and the 2 charts of the account 
  const [showButtonContent, setShowButtonContent] = useState("Show details");
  const [transactionTablePresent, setTransactionTablePresent] = useState(false);
  const [accountTableObject, setAccountTableObject] = useState({
    title: null,
    transactionCount: 0,
    transactionList: [],
  });

  // state of the form
  const [updateFormPresent, setUpdateFormPresent] = useState(false);
  const [deleteFormPresent, setDeleteFormPresent] = useState(false);

  // position elements of the card depending on the card's type 
  const topSize = type === "Debit" ? "60px" : "10px";
  const isCreditAccount = type !== "Debit";

  // Fetch data about the expense change and composition from API 
  useEffect(() => {
    expenseappClient.get(`/accounts/${accountInfo.id}/summary`)
      .then((response) => {
        // set the expense change and expense composition
        setExpenseChange(response.data.change_percentage);
        setExpenseComposition(response.data.composition_percentage);
      })
      .catch((error) => handleError(error));
  }, []);

  /**
   * Call the correct API request to fetch the transactions data 
   */
  function transactionListServer(transactionType = "latest", category = null) {
    if (transactionType === "latest") {
      return expenseappClient.get(`accounts/${accountInfo.id}/transactions`);
    } else {
      return expenseappClient.get(`accounts/${accountInfo.id}/transactions/both?category=${category}`);
    }
  }

  /**
   * Handle the category click action
   */
  function handleCategoryClick(e, argChartRef) {
    const categoryList = ["Housing", "Automobile", "Medical", "Subscription", "Grocery", "Dining", "Shopping", "Gas", "Others"];
    const selectedElement = getElementAtEvent(argChartRef.current, e)[0];
 
    if (selectedElement) {
      const selectedCategory = categoryList[selectedElement.index];
      transactionListServer("category", selectedCategory)
        .then((response) => {
          setAccountTableObject({
            title: newCategory,
            transactionCount: response.data.count,
            transactionList: response.data.results,
          });
        })
        .catch((error) => handleError(error));
    }
  }

  /**
   * Reset the original list of acount's transaction 
   */
  function fetchOriginalTransactions() {
    transactionListServer()
      .then((response) => {
        setAccountTableObject({
          title: "Latest",
          transactionCount: response.data.count,
          transactionList: response.data.results,
        });
      })
      .catch((error) => handleError(error));
  }

  /**
   * Handle the click to show the transaction table 
   */
  function showTableClick() {
    if (!transactionTablePresent) {
      setShowButtonContent("Hide details");
      setTransactionTablePresent(true);
      // reset the transaction table object when show it again 
      fetchOriginalTransactions();
    }
    else {
      setShowButtonContent("Show details");
      setTransactionTablePresent(false);
    }
  }

  return (
    <div className="account-detail-card">
      <div className="account-info" >
        <div className="institution-name">{accountData.institution}</div>
        <div className="account-name-num">{accountData.name}
          <span className="account-num">
            (****{accountData.account_number})
          </span>
        </div>
        {type === "Credit" && (
          <>
            {/* the info strictly for the credit accounts */}
            <div className="due-date">
              <span style={{ fontSize: "15px", fontWeight: "normal" }}>Due date: </span>
              {accountData.due_date}
            </div>
            <div className="credit-limit">
              <span style={{ fontSize: "15px", fontWeight: "normal" }}>Credit limit: </span>
              ${accountData.credit_limit}
            </div>
          </>
        )}
        <div className="balance" style={{ position: "relative", top: topSize }}>
          <span style={{ fontSize: "15px", fontWeight: "normal" }}>Balance: </span>${accountData.balance}
        </div>
      </div>
      <div className="button-list">
        {/* The button to show details */}
        <button className="show-detail-button" onClick={showTableClick}>{showButtonContent}</button>

        {/* The button to update the account */}
        <button className="update-info-button" onClick={() => setUpdateFormPresent(true)}>Update account</button>

        {/* The button to delete the account */}
        <button className="delete-button" onClick={() => setDeleteFormPresent(true)}>Delete account</button>
      </div>
      {transactionTablePresent && (
        <div style={{ marginTop: "20px", textAlign: "center" }}>
          <div className="charts">
            <div>
              <ChangeChart changeObject={expenseChange} />
            </div>
            <div>
              <CompositionChart
                compositionObject={expenseComposition}
                handleCategoryClick={handleCategoryClick} />
            </div>
          </div>
          <p style={{ 
            color: "darkcyan", 
            fontSize: "18px", 
            fontWeight: "bold" 
          }}>
            Click on category in the pie chart to show the account's transactions
          </p>
          {/* if the table's type is category, shows the go back button */}
          {accountTableObject.title !== "Latest" && (
            <button className="go-back-button"
              onClick={fetchOriginalTransactions}>Go back to latest list</button>
          )}
          <TransactionTable
            listName={"Account's " + accountTableObject.title + " Transactions"}
            transactionCount={accountTableObject.transactionCount}
            transactionList={accountTableObject.transactionList}
          />
        </div>
      )}
      {updateFormPresent && (
        <>
          <h4 style={{ textAlign: "center" }}>Update the info you need</h4>
          <AccountForm
            isCreditAccount={isCreditAccount}
            type="UPDATE"
            handleHideForm={() => setUpdateFormPresent(false)}
            handleChangeAccInfo={(newAccountInfo) => setAccountData({ ...newAccountInfo })}
            currentData={{
              ...accountData,
              due_date: reformatDate(accountInfo.due_date, '/', '-'),
            }} />
        </>
      )}
      {deleteFormPresent && (<DeleteAccountForm
        accountId={accountData.id}
        accountType={type}
        handleChangeAccList={handleChangeAccList}
        handleHideForm={() => setDeleteFormPresent(false)}
      />)}
    </div>
  );
}