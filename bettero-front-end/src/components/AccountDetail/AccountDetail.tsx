import { useState, useEffect, MouseEvent } from 'react';
import handleError from '@provider/handleError';
import TransactionTable from '@components/TransactionTable/TransactionTable';
import DeleteAccountForm from '@Forms/DeleteForms/DeleteAccountForm';
import AccountForm from '@Forms/AddForms/AccountForm';
import ChangeChart from '@Charts/ChangeChart';
import CompositionChart from '@Charts/CompositionChart';
import { getElementAtEvent } from 'react-chartjs-2';
import { reformatDate } from '@utils';
import { getAccountSummary, getAccountTransactions } from '@provider/api';
import { Account, CategoryObject, Transaction } from '@interface';
import './AccountDetail.scss';

interface AccountDetailProps {
  type: string,
  accountInfo: Account, 
  onChangeAccList: (data: Account[]) => void
}

export default function AccountDetail({ type = "Debit", accountInfo, onChangeAccList }: AccountDetailProps) {
  // info of the account 
  const [accountData, setAccountData] = useState<Account>({ ...accountInfo });

  // the expense change and composition data for the charts 
  const [expenseChange, setExpenseChange] = useState<CategoryObject>();
  const [expenseComposition, setExpenseComposition] = useState<CategoryObject>();

  // state of the transaction table and the 2 charts of the account 
  const [showButtonContent, setShowButtonContent] = useState("Show details");
  const [transactionTablePresent, setTransactionTablePresent] = useState(false);
  const [accountTableObject, setAccountTableObject] = useState({
    title: 'Latest',
    transactionCount: 0,
    transactionList: [] as Transaction[],
  });

  // state of the form
  const [updateFormPresent, setUpdateFormPresent] = useState(false);
  const [deleteFormPresent, setDeleteFormPresent] = useState(false);

  // position elements of the card depending on the card's type 
  const topSize = type === "Debit" ? "60px" : "10px";
  const isCreditAccount = type !== "Debit";

  /**
   * Handle the category click action
   */
  function handleCategoryClick(e: MouseEvent<HTMLCanvasElement>, argChartRef: any): void {
    const categoryList = ["Housing", "Automobile", "Medical", "Subscription", "Grocery", "Dining", "Shopping", "Gas", "Others"];
    const selectedElement = getElementAtEvent(argChartRef.current, e)[0];
 
    if (selectedElement) {
      const selectedCategory = categoryList[selectedElement.index];
      getAccountTransactions(accountInfo.id, "category", selectedCategory)
        .then((response) => setAccountTableObject({ title: selectedCategory, ...response.data }))
        .catch((error) => handleError(error));
    }
  }

  /**
   * Reset the original list of acount's transaction 
   * @returns {void}
   */
  function fetchOriginalTransactions(): void {
    getAccountTransactions(accountInfo.id)
      .then((response) => setAccountTableObject({title: "Latest", ...response.data}))
      .catch((error) => handleError(error));
  }

  /**
   * Fetch the page of list of transactions based on the page index 
   * @param {number} pageIndex 
   * @returns {void}
   */
  function handleNextClick(pageIndex: number): void {
    let transactionType: string; 
    let category: string | undefined;
    if (accountTableObject.title === 'Latest') {
      transactionType = 'latest'; 
      category = undefined;
    } else {
      transactionType = accountTableObject.title; 
      category = accountTableObject.title; 
    }

    getAccountTransactions(accountInfo.id, transactionType, category, pageIndex)
      .then(response => setAccountTableObject({title: accountTableObject.title, ...response.data}))
      .catch(error => handleError(error));
  }

  /**
   * Handle the click to show the transaction table 
   */
  function showTableClick(): void {
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

  // Fetch data about the expense change and composition from API 
  useEffect(() => {
    getAccountSummary(accountInfo.id)
      .then((response) => {
        // set the expense change and expense composition
        setExpenseChange(response.data.changePercentage);
        setExpenseComposition(response.data.compositionPercentage);
      })
      .catch((error) => handleError(error));
  }, []);

  return (
    <div className="account-detail-card">
      <div className="account-info" >
        <div className="institution-name">{accountData.institution}</div>
        <div className="account-name-num">
          {accountData.name}
          <span className="account-num">(****{accountData.accountNumber})</span>
        </div>
        {type === "Credit" && 
          <>
            {/* The info strictly for the credit accounts */}
            <div className="due-date">
              <span style={{ fontSize: "15px", fontWeight: "normal" }}>Due date: </span>
              {accountData.dueDate!.toString()}
            </div>
            <div className="credit-limit">
              <span style={{ fontSize: "15px", fontWeight: "normal" }}>Credit limit: </span>
              ${accountData.creditLimit}
            </div>
          </>
        }
        <div className="balance" style={{ position: "relative", top: topSize }}>
          <span style={{ fontSize: "15px", fontWeight: "normal" }}>Balance: </span>${accountData.balance}
        </div>
      </div>
      <div className="button-list">
        {/* The button to show details */}
        <button className="show-detail-button" onClick={showTableClick}>{showButtonContent}</button>

        {/* The button to update the account */}
        <button className="update-info-button" onClick={() => setUpdateFormPresent(true)}>
          Update account
        </button>

        {/* The button to delete the account */}
        <button className="delete-button" onClick={() => setDeleteFormPresent(true)}>
          Delete account
        </button>
      </div>
      {transactionTablePresent && (
        <div style={{ marginTop: "20px", textAlign: "center" }}>
          <div className="charts">
            <div>
              <ChangeChart intervalType="month" changeObject={expenseChange!} />
            </div>
            <div>
              <CompositionChart
                intervalType='month' chartType='regular'
                compositionObject={expenseComposition!} onCategoryClick={handleCategoryClick} />
            </div>
          </div>
          <p style={{ color: "darkcyan", fontSize: "18px", fontWeight: "bold" }}>
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
            onChangeTransList={handleNextClick}
          />
        </div>
      )}
      {updateFormPresent && (
        <>
          <h4 style={{ textAlign: "center" }}>Update the info you need</h4>
          <AccountForm
            type="UPDATE" isCreditAccount={isCreditAccount}
            currentData={{
              ...accountData,
              dueDate: accountInfo.dueDate ? 
                new Date(reformatDate(accountInfo.dueDate?.toString(), '/', '-')) : null,
            }}
            onHide={() => setUpdateFormPresent(false)}
            onChangeAccInfo={(newAccountInfo: Account) => setAccountData({ ...newAccountInfo })}
          />
        </>
      )}
      {deleteFormPresent && <DeleteAccountForm
        accountId={accountData.id} accountType={type}
        onChangeAccList={onChangeAccList} onHide={() => setDeleteFormPresent(false)}
      />}
    </div>
  );
}