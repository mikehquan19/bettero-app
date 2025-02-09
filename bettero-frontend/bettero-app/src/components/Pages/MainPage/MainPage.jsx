import { expenseappClient } from '../../../provider/api.js';
import handleError from '../../../provider/handleError.js';
import Card from '../../Card/Card.jsx';
import TransactionTable from '../../TransactionTable/TransactionTable.jsx';
import AccountForm from '../../Forms/AddForms/AccountForm.jsx';
import TransactionForm from '../../Forms/AddForms/TransactionForm.jsx';
import ChangeChart from '../../Charts/ChangeChart.jsx';
import CompositionChart from '../../Charts/CompositionChart.jsx';
import TimeseriesChart from '../../Charts/TimeseriesChart.jsx';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faCirclePlus } from '@fortawesome/free-solid-svg-icons'
import { useEffect, useState } from 'react';
import { getElementAtEvent } from 'react-chartjs-2';
import { Link } from 'react-router-dom';
import './MainPage.css';

// the FinancialCards component 
const FinancialCards = ({ financialData }) => {
  return (
    <>
      <h2 className="financial-info-title">This month's financial information</h2>
      <div className="financial-cards-wrapper">
        <div className="financial-card">
          <h4>Total balance: </h4>
          <div>${financialData.total_balance}</div>
        </div>
        <div className="financial-card">
          <h4>Amount due: </h4>
          <div>${financialData.total_amount_due}</div>
        </div>
        <div className="financial-card">
          <h4>Total income: </h4>
          <div>${financialData.total_income}</div>
        </div>
        <div className="financial-card">
          <h4>Total expense: </h4>
          <div>${financialData.total_expense}</div>
        </div>
      </div>
    </>
  );
}


// the Button component
const Button = ({ content, id = null, handleEvents = null }) => {
  return (
    <div style={{textAlign: "center"}}>
      <button className="add-button" id={id}
      onMouseOver={(e) => {
        e.target.style.backgroundColor = "white";
        e.target.style.border = "2px solid darkcyan";
        e.target.style.color = "darkcyan";
      }}
      onMouseOut={(e) => {
        e.target.style.backgroundColor = "darkcyan";
        e.target.style.border = "1px solid green";
        e.target.style.color = "white";
      }}
      onClick={handleEvents}> <FontAwesomeIcon icon={faCirclePlus} style={{
        marginRight: "5px",
        display: "block",
        margin: "0 auto 5px auto",
      }} /> {content} </button>
    </div>
  );
}


// the CardFrame component, wrapping around the card 
const CardFrame = ({ children, cardType, numCards }) => {
  return (
    <div className="card-frame">
      <h3>
        {/* go to the detail page */}
        <Link to={`/accounts/${cardType}`}
          style={{ color: "black", }}
          onMouseOver={(e) => { e.target.style.color = "white"; }}
          onMouseOut={(e) => { e.target.style.color = "black"; }}>
          {cardType.charAt(0).toUpperCase() + cardType.slice(1)} Accounts ({numCards})
        </Link>
      </h3>
      <div className="card-list">
        {/* render the children */}
        {children}
      </div>
    </div>
  );
}


// export the App component 
const MainPage = ({ navbarWidth, titleHeight }) => {

  // data for charts 
  const [financialInfo, setFinancialInfo] = useState({});
  const [expenseChange, setExpenseChange] = useState({});
  const [expenseComposition, setExpenseComposition] = useState({});
  const [timeseriesData, setTimeseriesData] = useState({});

  // list of credit and debit accounts 
  const [accountList, setAccountList] = useState([]);

  // if the category table is displayed 
  // the content of the categorical transaction table and latest transaction table
  const [latestTransList, setLatestTransList] = useState([]);
  const [tablePresent, setTablePresent] = useState(false);
  const [tableCategoryObject, setTableCategoryObject] = useState({
    category: "Grocery",
    categoryTransList: [],
  });

  // fetch data from backend 
  const fetchData = () => {
    Promise.all([
      expenseappClient.get("/transactions"),
      expenseappClient.get("/summary"),
      expenseappClient.get("/accounts")
    ])
      // when all the requests have been fullfilled 
      .then((response) => {
        setLatestTransList(response[0].data);
        // set the financial info
        const summaryData = response[1].data;
        setFinancialInfo({
          total_balance: summaryData.total_balance.toFixed(2),
          total_amount_due: summaryData.total_amount_due.toFixed(2),
          total_income: summaryData.total_income.toFixed(2),
          total_expense: summaryData.total_expense.toFixed(2),
        });

        // set the expense change, composition, and timeseries data 
        setExpenseChange(summaryData.change_percentage);
        setExpenseComposition(summaryData.composition_percentage);
        setTimeseriesData(summaryData.daily_expense);

        // set the list of debit and credit accounts
        setAccountList(response[2].data);
      })
      // if there's any error 
      .catch((error) => handleError(error));
  }
  useEffect(fetchData, []);

  
  // state of the form to add debit account, credit account, transaction
  const [debitFormPresent, setDebitFormPresent] = useState(false);
  const [creditFormPresent, setCreditFormPresent] = useState(false);
  const [transFormPresent, setTransFormPresent] = useState(false);

  // list of expense categories 
  const categoryList = ["Housing", "Automobile", "Medical", "Subscription", "Grocery", "Dining", "Shopping", "Gas", "Others"];

  // same callback function 
  // always display the table and change the content of table based on the button
  const handleCategoryClick = (e, argChartRef) => {
    const selectedElement = getElementAtEvent(argChartRef.current, e)[0];
    // if the selected element is existent 
    if (selectedElement) {
      setTablePresent(true); // always displays the table 

      // set the category to the picked one 
      const newCategory = categoryList[selectedElement.index];
      expenseappClient.get(`/transactions/category/${newCategory}`)
        .then((response) => {
          setTableCategoryObject({
            category: newCategory,
            categoryTransList: response.data,
          })
        })
        .catch((error) => handleError(error));
    }
  }

  // always hide the table 
  // callback function 
  const handleHideClick = () => {
    setTablePresent(false);
  }

  // display or hide the debit account form 
  // callback passed to AccountForm component
  const handleDebitPresentClick = () => {
    setDebitFormPresent(!debitFormPresent);
  }

  // display or hide the credit account form 
  const handleCreditPresentClick = () => {
    setCreditFormPresent(!creditFormPresent);
  }

  // display or hide the transaction form 
  const handleTransFormPresentClick = () => {
    setTransFormPresent(!transFormPresent);
  }

  // list of debit and credit accounts 
  const debitAccountList = accountList.filter(account => account.account_type === "Debit");
  const creditAccountList = accountList.filter(account => account.account_type === "Credit");

  // render the page 
  return (
    <div style={{
      marginTop: `${titleHeight}px`,
      marginLeft: `${navbarWidth}px`,
      width: `calc(100% - ${navbarWidth}px)`,
    }}>
      <FinancialCards financialData={financialInfo} />

      {/* the list of buttons */}
      <div className="charts">
        <div className="main-charts">
          <div className="timeseries-chart-wrapper">
            <TimeseriesChart timeseriesObject={timeseriesData} />
          </div>
          <div className="chart-wrapper">
            <div className="change-chart-wrapper">
              <ChangeChart changeObject={expenseChange} />
            </div>
            <div className="composition-chart-wrapper">
              <CompositionChart 
                compositionObject={expenseComposition} handleCategoryClick={handleCategoryClick} />
            </div>
          </div>
          <p>Click on the category in the pie to show list of its transactions</p>
        </div>
        {tablePresent && (
          <>
            <TransactionTable
              listName={tableCategoryObject.category !== "Income" ?
                tableCategoryObject.category + " expenses this month" :
                tableCategoryObject.category + " incomes this month"}
              transactionList={tableCategoryObject.categoryTransList} />

            <Button content="Hide list" handleEvents={handleHideClick} />
          </>
        )}
      </div>

      <CardFrame cardType="debit" numCards={debitAccountList.length}>
        {/* The list of debit accounts */}
        {debitAccountList.length === 0 && (
          <p style={{fontWeight: "bold", margin: "0 auto"}}>There are no debit accounts</p>
        )}
        {debitAccountList.map(debitAccount =>
          <Card key={debitAccount.id} accountInfo={debitAccount} />
        )}
      </CardFrame>

      {!debitFormPresent && 
        <Button content="Add Account" handleEvents={handleDebitPresentClick} />}

      {debitFormPresent &&
        (<AccountForm
          handleChangeAccList={(newAccountList) => setAccountList(newAccountList)}
          isCreditAccount={false} handleHideForm={handleDebitPresentClick} />)}
      <br />

      <CardFrame cardType="credit" numCards={creditAccountList.length}>
        {/* the list of credit accounts */}
        {creditAccountList.length === 0 && (
          <p style={{fontWeight: "bold", fontSize:"1.2rem", margin: "0 auto"}}>There are no credit accounts</p>
        )}
        {creditAccountList.map(creditAccount =>
          <Card key={creditAccount.id} accountInfo={creditAccount} type="credit" />
        )}

      </CardFrame>
      {!creditFormPresent && 
        <Button content="Add Account" handleEvents={handleCreditPresentClick} />}

      {/* Form to add credit accounts */}
      {creditFormPresent &&
        (<AccountForm
          handleChangeAccList={(newAccountList) => setAccountList(newAccountList)}
          isCreditAccount={true} handleHideForm={handleCreditPresentClick} />)}

      <TransactionTable
        listName="Latest transactions" transactionList={latestTransList} />
      {!transFormPresent && <Button content="Add Transaction" handleEvents={handleTransFormPresentClick} />}

      {/* Form to add transactions */}
      {transFormPresent &&
        (<TransactionForm
          accountListData={accountList} handleChangeTransList={fetchData}
          handleHideForm={handleTransFormPresentClick} />)}
    </div>
  );
}

export default MainPage; 