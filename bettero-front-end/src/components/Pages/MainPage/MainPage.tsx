import handleError from '@provider/handleError';
import FinancialCard from '@components/FinancialCard/FinancialCard';
import Button from '@components/Button/Button';
import Card from '@components/Card/Card';
import CardFrame from '@components/CardFrame/CardFrame';
import TransactionTable from '@components/TransactionTable/TransactionTable';
import AccountForm from '@Forms/AddForms/AccountForm';
import TransactionForm from '@Forms/AddForms/TransactionForm';
import ChangeChart from '@Charts/ChangeChart';
import CompositionChart from '@Charts/CompositionChart';
import TimeseriesChart from '@Charts/TimeseriesChart';
import { getUserAccounts, getUserTransactions, getFinancialSummary, getCategoryTransactions} from '@provider/api';
import { useEffect, useState, MouseEvent } from 'react';
import { getElementAtEvent } from 'react-chartjs-2';
import { Account, CategoryObject, FinancialInfo, PageProps, Transaction } from '@interface';
import './MainPage.scss';

/**
 * The Main Page
 * @param {PageProps} {navbarWidth, titleHeight}
 * @returns {JSX.Element}
 */
export default function MainPage({ navbarWidth, titleHeight }: PageProps): JSX.Element {

  const [isLoading, setIsLoading] = useState(true);
  const [financialInfo, setFinancialInfo] = useState<FinancialInfo>();

  // Data for charts 
  const [expenseChange, setExpenseChange] = useState<CategoryObject>({} as CategoryObject);
  const [expenseComposition, setExpenseComposition] = useState<CategoryObject>();
  const [timeseriesData, setTimeseriesData] = useState<Record<string, number>>({});

  const [debitAccountList, setDebitAccountList] = useState<Account[]>([]);
  const [creditAccountList, setCreditAccountList] = useState<Account[]>([]);

  const [categoryTablePresent, setCategoryTablePresent] = useState<boolean>(false);
  // The content of the latest transaction table
  const [latestTableObj, setLatestTableObj] = useState({
    page: 1,
    transactionCount: 0, 
    transactionList: [] as Transaction[]
  });
  // The content of the categorical transaction table 
  const [categoryTableObj, setCategoryTableObj] = useState({
    category: "Grocery",
    page: 1,
    transactionCount: 0, 
    transactionList: [] as Transaction[],
  });

  // state of the form to add debit account, credit account, transaction
  const [debitFormPresent, setDebitFormPresent] = useState(false);
  const [creditFormPresent, setCreditFormPresent] = useState(false);
  const [transactionFormPresent, setTransactionFormPresent] = useState(false);

  /**
   * Fetch data from backend 
   */
  function fetchData(): void {
    Promise.all([getFinancialSummary(), getUserAccounts(), getUserTransactions()])
      .then((response) => {
        // set the financial info
        setFinancialInfo(response[0].data.financialInfo);
        setExpenseChange(response[0].data.changePercentage);
        setExpenseComposition(response[0].data.compositionPercentage);
        setTimeseriesData(response[0].data.dailyExpense);

        // set the list of debit and credit accounts
        const accountData = response[1].data as Account[]; 
        setDebitAccountList(accountData.filter(account => account.accountType === "Debit"));
        setCreditAccountList(accountData.filter(account => account.accountType === "Credit"));
 
        // set transaction object
        setLatestTableObj(response[2].data);
        setIsLoading(false);
      })
      .catch((error) => handleError(error));
  }
  useEffect(fetchData, []);

  /**
   * Display the table and change the content of table based on the button
   * @param {MouseEvent<HTMLCanvasElement>} e - Clicking event 
   * @param {any} argChartRef - Reference to the canvas on the chart
   */
  function handleCategoryClick(e: MouseEvent<HTMLCanvasElement>, argChartRef: any): void {
    const categoryList = ["Housing", "Automobile", "Medical", "Subscription", "Grocery", "Dining", "Shopping", "Gas", "Others"];
    const selectedElement = getElementAtEvent(argChartRef.current, e)[0];

    if (selectedElement) {
      // If the selected element is existent, always displays the table
      setCategoryTablePresent(true);

      getCategoryTransactions(categoryList[selectedElement.index]) // Set the category to the picked one 
        .then((response) => {
          setCategoryTableObj({
            category: categoryList[selectedElement.index], 
            ...response.data
          });
        })
        .catch((error) => handleError(error));
    }
  }

  /**
   * Handle action to click next for the latest transaction table
   * @param {number} pageIndex
   * @returns {void}
   */
  function handleNextClick(pageIndex: number): void {
    getUserTransactions(pageIndex)
      .then(response => setLatestTableObj(response.data))
      .catch(error => handleError(error));
  }

  function handleCategoryNextClick(pageIndex: number): void {
    getCategoryTransactions(categoryTableObj.category, pageIndex)
      .then(response => setCategoryTableObj({category: categoryTableObj.category, ...response.data}))
      .catch(error => handleError(error));
  }

  if (isLoading) return <div>Loading...</div>
  return (
    <div style={{ 
      marginTop: `${titleHeight}px`, 
      marginLeft: `${navbarWidth}px`, 
      width: `calc(100% - ${navbarWidth}px)` 
    }}>
      <FinancialCard financialData={financialInfo} />

      {/* the list of buttons */}
      <div className="charts">
        <div className="main-charts">
          <div className="timeseries-chart-wrapper">
            <TimeseriesChart contentType="expense" intervalType="month" timeSeriesObject={timeseriesData} />
          </div>
          <div className="chart-wrapper">
            <div className="change-chart-wrapper">
              <ChangeChart intervalType="month" changeObject={expenseChange!} />
            </div>
            <div className="composition-chart-wrapper">
              <CompositionChart 
                intervalType='month' chartType='regular'
                compositionObject={expenseComposition!} onCategoryClick={handleCategoryClick} />
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
              currentIndex={categoryTableObj.page}
              transactionCount={categoryTableObj.transactionCount} transactionList={categoryTableObj.transactionList}
              onChangeTransList={handleCategoryNextClick} />
            <Button 
              content="Hide list" 
              onClick={() => setCategoryTablePresent(false)} />
          </>
        }
      </div>

      {/* The list of debit accounts */}
      <CardFrame cardType="debit" numCards={debitAccountList.length}>
        {!debitAccountList.length && 
          <p style={{fontWeight: "bold", margin: "0 auto"}}>There are no debit accounts</p>
        }
        {debitAccountList.map(debitAccount => 
          <Card key={debitAccount.id} 
            accountInfo={debitAccount} theme={{ backgroundColor: 'blue', color: 'white' }} type="debit" />)
        }
      </CardFrame>
      <Button 
        content="Add Account" 
        onClick={() => setDebitFormPresent(!debitFormPresent)} />

      {/* Form to add debit accounts */}
      {debitFormPresent &&
        <AccountForm
          type="ADD" isCreditAccount={false}
          onChangeAccList={(newAccountList) => setDebitAccountList(newAccountList)} 
          onHide={() => setDebitFormPresent(!debitFormPresent)} />
      }

      <br/>

      {/* List of credit accounts */}
      <CardFrame cardType="credit" numCards={creditAccountList.length}>
        {!creditAccountList.length && 
          <p style={{fontWeight: "bold", margin: "0 auto"}}>There are no credit accounts</p>
        }
        {creditAccountList.map(creditAccount => 
          <Card key={creditAccount.id} 
            accountInfo={creditAccount} theme={{ backgroundColor: 'blue', color: 'white' }} type="credit" />
        )}
      </CardFrame>
      <Button 
        content="Add Account" 
        onClick={() => setCreditFormPresent(!creditFormPresent)} />

      {/* Form to add credit accounts */}
      {creditFormPresent &&
        <AccountForm
          type="ADD" isCreditAccount={true} 
          onChangeAccList={(newAccountList) => setCreditAccountList(newAccountList)}
          onHide={() => setCreditFormPresent(!creditFormPresent)} />
      }

      <TransactionTable
        listName="Latest transactions" 
        currentIndex={latestTableObj.page}
        transactionCount={latestTableObj.transactionCount} transactionList={latestTableObj.transactionList}
        onChangeTransList={handleNextClick} />
      <Button 
        content="Add Transaction" 
        onClick={() => setTransactionFormPresent(!transactionFormPresent)} />

      {/* Form to add transactions */}
      {transactionFormPresent &&
        <TransactionForm
          accountList={[...debitAccountList, ...creditAccountList]} 
          onChangeTransList={(data) => setLatestTableObj(data)} 
          onHide={() => setTransactionFormPresent(!transactionFormPresent)} />
      }
    </div>
  );
}