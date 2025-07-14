import { expenseappClient } from '../../../provider/api.js';
import handleError from '../../../provider/handleError.js';
import FinancialCard from '../../FinancialCard/FinancialCard.jsx';
import Button from '../../Button/Button.jsx';
import Card from '../../Card/Card.jsx';
import CardFrame from '../../CardFrame/CardFrame.jsx';
import TransactionTable from '../../TransactionTable/TransactionTable.jsx';
import AccountForm from '../../Forms/AddForms/AccountForm.jsx';
import TransactionForm from '../../Forms/AddForms/TransactionForm.jsx';
import ChangeChart from '../../Charts/ChangeChart.jsx';
import CompositionChart from '../../Charts/CompositionChart.jsx';
import TimeseriesChart from '../../Charts/TimeseriesChart.jsx';
import { useEffect, useState } from 'react';
import { getElementAtEvent } from 'react-chartjs-2';
import './MainPage.css';


function MainPage({ navbarWidth, titleHeight }) {

  // Data for charts 
  const [financialInfo, setFinancialInfo] = useState({});
  const [expenseChange, setExpenseChange] = useState({});
  const [expenseComposition, setExpenseComposition] = useState({});
  const [timeseriesData, setTimeseriesData] = useState({});
  const [debitAccountList, setDebitAccountList] = useState([]);
  const [creditAccountList, setCreditAccountList] = useState([]);

  // Whether the category table is present 
  const [categoryTablePresent, setCategoryTablePresent] = useState(false);
  // The content of the latest transaction table
  const [latestTableObj, setLatestTableObj] = useState({
    previousPage: null, 
    nextPage: null, 
    transactionCount: 0, 
    transactionList: []
  });
  // The content of the categorical transaction table 
  const [categoryTableObj, setCategoryTableObj] = useState({
    category: "Grocery",
    previousPage: null, 
    nextPage: null, 
    transactionCount: 0, 
    transactionList: [],
  });

  // state of the form to add debit account, credit account, transaction
  const [debitFormPresent, setDebitFormPresent] = useState(false);
  const [creditFormPresent, setCreditFormPresent] = useState(false);
  const [transactionFormPresent, setTransactionFormPresent] = useState(false);

  /**
   * Get the appropriate promise based on the given type 
   */
  function initialDataServer() {
    return Promise.all([
      expenseappClient.get('/transactions'), expenseappClient.get('/summary'), expenseappClient('/accounts')
    ]);
  }

  function transactionListServer(selectedCategory) {
    return expenseappClient.get(`/transactions/category/${selectedCategory}`);
  }

  // Fetch data from backend 
  function fetchData() {
    initialDataServer()
      .then((response) => {
        setLatestTableObj((previousData) => ({
          ...previousData,
          transactionCount: response[0].data.count, 
          transactionList: response[0].data.results
        }));

        // set the financial info
        const summaryData = response[1].data;
        setFinancialInfo({
          totalBalance: summaryData.total_balance,
          totalAmountDue: summaryData.total_amount_due,
          totalIncome: summaryData.total_income,
          totalExpense: summaryData.total_expense
        });

        // set the expense change, composition, and timeseries data 
        setExpenseChange(summaryData.change_percentage);
        setExpenseComposition(summaryData.composition_percentage);
        setTimeseriesData(summaryData.daily_expense);

        // set the list of debit and credit accounts
        setDebitAccountList(response[2].data.filter((account) => account.type === "Debit"));
        setCreditAccountList(response[2].data.filter((account) => account.type === "Credit"));

      })
      .catch((error) => handleError(error));
  }
  useEffect(fetchData, []);

  // Display the table and change the content of table based on the button
  function handleCategoryClick(e, argChartRef) {
    const categoryList = [
      "Housing", "Automobile", "Medical", "Subscription", "Grocery", "Dining", "Shopping", "Gas", "Others"
    ];
    const selectedElement = getElementAtEvent(argChartRef.current, e)[0];

    if (selectedElement) {
      setCategoryTablePresent(true); // If the selected element is existent, always displays the table

      // Set the category to the picked one 
      transactionListServer(categoryList[selectedElement.index])
        .then((response) => {
          setCategoryTableObj({
            category: categoryList[selectedElement.index],
            transactionCount: response.data.count,
            transactionList: response.data.results,
          });
        })
        .catch((error) => handleError(error));
    }
  }

  return (
    <div style={{ marginTop: `${titleHeight}px`, marginLeft: `${navbarWidth}px`, width: `calc(100% - ${navbarWidth}px)` }}>
      <FinancialCard financialData={financialInfo} />

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
        {categoryTablePresent && 
          <>
            <TransactionTable
              listName={categoryTableObj.category !== "Income" ? 
                categoryTableObj.category + " expenses this month" : 
                categoryTableObj.category + " incomes this month"
              }
              transactionCount={categoryTableObj.transactionCount} transactionList={categoryTableObj.transactionList} />
            <Button 
              content="Hide list" 
              handleEvents={() => setCategoryTablePresent(false)} />
          </>
        }
      </div>

      {/* The list of debit accounts */}
      <CardFrame cardType="debit" numCards={debitAccountList.length}>
        {debitAccountList.length === 0 && 
          <p style={{fontWeight: "bold", margin: "0 auto"}}>There are no debit accounts</p>
        }
        {debitAccountList.map(debitAccount => 
          <Card key={debitAccount.id} accountInfo={debitAccount} />)
        }
      </CardFrame>
      {!debitFormPresent && 
        <Button 
          content="Add Account" 
          handleEvents={() => setDebitFormPresent(!debitFormPresent)} />
      }

      {/* Form to add debit accounts */}
      {debitFormPresent &&
        <AccountForm
          handleChangeAccList={(newAccountList) => setDebitAccountList(newAccountList)}
          isCreditAccount={false} 
          handleHideForm={() => setDebitFormPresent(!debitFormPresent)} />
      }
      <br/>

      {/* List of credit accounts */}
      <CardFrame cardType="credit" numCards={creditAccountList.length}>
        {creditAccountList.length === 0 && 
          <p style={{fontWeight: "bold", margin: "0 auto"}}>There are no credit accounts</p>
        }
        {creditAccountList.map(creditAccount => 
          <Card key={creditAccount.id} accountInfo={creditAccount} type="credit" />
        )}
      </CardFrame>
      {!creditFormPresent && 
        <Button 
          content="Add Account" 
          handleEvents={() => setCreditFormPresent(!creditFormPresent)} />
      }

      {/* Form to add credit accounts */}
      {creditFormPresent &&
        <AccountForm
          handleChangeAccList={(newAccountList) => setCreditAccountList(newAccountList)}
          isCreditAccount={true} 
          handleHideForm={() => setCreditFormPresent(!creditFormPresent)} />
      }

      <TransactionTable
        listName="Latest transactions" 
        transactionCount={latestTableObj.transactionCount} transactionList={latestTableObj.transactionList} />
      {!transactionFormPresent && 
        <Button 
          content="Add Transaction" 
          handleEvents={() => setTransactionFormPresent(!transactionFormPresent)} />
      }

      {/* Form to add transactions */}
      {transactionFormPresent &&
        <TransactionForm
          accountListData={accountList} 
          handleChangeTransList={fetchData} 
          handleHideForm={() => setTransactionFormPresent(!transactionFormPresent)} />
      }
    </div>
  );
}

export default MainPage; 