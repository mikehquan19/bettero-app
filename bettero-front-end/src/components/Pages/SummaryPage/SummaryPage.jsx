import { useEffect, useState } from 'react';
import { expenseappClient } from '../../../provider/api.js';
import handleError from '../../../provider/handleError.js';
import ChangeChart from '../../Charts/ChangeChart.jsx';
import CompositionChart from '../../Charts/CompositionChart.jsx';
import TimeseriesChart from '../../Charts/TimeseriesChart.jsx';
import ExpenseChart from '../../Charts/ExpenseChart.jsx';
import TransactionTable from '../../TransactionTable/TransactionTable.jsx';
import { getElementAtEvent } from 'react-chartjs-2';
import { useMediaQuery } from '@uidotdev/usehooks';
import { reformatDate } from '../../../utils.js';
import { latestIntervals, latestIntervalExpense, latestIntervalChart } from '../../../utils.js';
import './SummaryPage.css';

function SummaryPage({ navbarWidth, titleHeight }) {
  const isMobileDevice = useMediaQuery("only screen and (max-width : 500px)");

  // Latest intervals and total expense of each interval
  const [intervalObject, setIntervalObject] = useState({ "month": [], "bi-week": [], "week": [] });
  const [isLoading, setIsLoading] = useState(true);

  const [intervalType, setIntervalType] = useState("Month");
  const [clickedInterval, setClickedInterval] = useState(0);
  const [intervalTableObj, setIntervalTableObj] = useState({
    title: "All",
    transactionCount: 0,
    transactionList: [],
  });

  /**
   * Get the appropriate promise based on the type 
   */
  function financialSummaryServer() {
    return Promise.all([expenseappClient.get('/full_summary'), expenseappClient.get('/transactions/initial')]);
  }

  function transactionListServer(serverType, category = null, firstDate = null, lastDate = null) {
    if (serverType === "intervalTransaction") {
      return expenseappClient.get(
        `/transactions/interval?first_date=${firstDate}&last_date=${lastDate}`
      );
    } else if (serverType === "bothTransaction") {
      return expenseappClient.get(
        `/transactions/both?category=${category}&first_date=${firstDate}&last_date=${lastDate}`
      );
    } else {
      throw new Error("Server type invalid!");
    }
  }

  /** 
   * Fetch the financial summary data 
   */
  function fetchFinancialSummary() {
    financialSummaryServer()
      .then((response) => {
        setIntervalObject(response[0].data.latest_interval_expense);
        setIntervalTableObj((previousData) => ({
          ...previousData,
          transactionCount: response[1].data.count,
          transactionList: response[1].data.results,
        }))
        setIsLoading(false);
      })
      .catch((error) => handleError(error));
  }

  /** 
   * Fetch data about transactions during the period 
   */
  function fetchPeriodTransactions() {
    const latestIntervalArr = latestIntervals(intervalObject, intervalType);

    if (latestIntervalArr.length !== 0) {
      const firstDate = latestIntervalArr[clickedInterval].first_date;
      const lastDate = latestIntervalArr[clickedInterval].last_date;

      transactionListServer("intervalTransaction", null, firstDate, lastDate)
        .then((response) => {
          setIntervalTableObj({
            title: "All",
            transactionCount: response.data.count,
            transactionList: response.data.results
          });
        })
        .catch(error => handleError(error));
    }
  }

  /**
   * Fetch data about interval transactions based on click 
   */
  function onIntervalClick(intervalIndex, firstDate, lastDate) {
    if (intervalIndex !== clickedInterval) {
      setClickedInterval(intervalIndex);
    } else {
      transactionListServer("intervalTransaction", null, firstDate, lastDate)
        .then((response) => {
          setIntervalTableObj({
            title: "All", 
            transactionCount: response.data.count,
            transactionList: response.data.results
          });
        })
        .catch((error) => handleError(error));
    }
  }

  /**
   * Fetch data of categorical transactions based on the click
   */
  function onCategoryClick(e, argChartRef, firstDate, lastDate) {
    const categoryList = [
      "Housing", "Automobile", "Medical", "Subscription", "Grocery", "Dining", "Shopping", "Gas", "Others"
    ];
    const selectedElement = getElementAtEvent(argChartRef.current, e)[0];

    if (selectedElement) {
      const selectedCategory = categoryList[selectedElement.index];
      transactionListServer("bothTransaction", selectedCategory, firstDate, lastDate)
        .then((response) => {
          setIntervalTableObj({
            title: selectedCategory,
            transactionCount: response.data.count,
            transactionList: response.data.results,
          });
        })
        .catch((error) => handleError(error));
    }
  }

  /**
   * Reset the interval type, which in turn will reset the data 
   */
  function resetInterval(intervalType) {
    setIntervalType(intervalType);
    setClickedInterval(0);
  }

  function setButtonColor(interval) {
    if (intervalType === interval) {
      return { backgroundColor: "rgba(0, 139, 139, 1)" };
    } else {
      return { backgroundColor: "rgba(0, 139, 139, 0)" };
    }
  }

  useEffect(() => fetchFinancialSummary(), []);
  // Only call after re-rendering triggered by interval type and clicked interval
  useEffect(() => fetchPeriodTransactions(), [intervalType, clickedInterval]);

  if (!isLoading) {
    // Type of interval to be displayed 
    const displayedType = intervalType === "Bi_week" ? "bi-week" : intervalType.toLowerCase();

    // Update the data about total expense of 5 latest intervals 
    const latestMonthExpense = latestIntervalExpense(intervalObject, "Month");
    const latestBiWeekExpense = latestIntervalExpense(intervalObject, "Bi_week");
    const latestWeekExpense = latestIntervalExpense(intervalObject, "Week");

    // Update the interval data based on the state  
    const latestIntervalArr = latestIntervals(intervalObject, intervalType);

    // Clicked first and last date
    const selectedFirstDate = latestIntervalArr[clickedInterval].first_date;
    const selectedLastDate = latestIntervalArr[clickedInterval].last_date;

    // Object for change and composition data of the clicked interval
    const intervalChartObject = latestIntervalChart(intervalObject, intervalType, selectedFirstDate);

    // Update the data for daily expense, change, and composition chart of that interval 
    const intervalDailyExpense = intervalChartObject.daily_expense;
    const intervalChange = intervalChartObject.change_percentage;
    const intervalComposition = intervalChartObject.composition_percentage;

    return (
      <div style={{
        marginTop: `${titleHeight}px`,
        marginLeft: `${navbarWidth}px`,
        width: `calc(100% - ${navbarWidth}px)`,
      }}>
        <h2 style={{
          width: "85%",
          margin: "0 auto",
        }}>Detailed summary tracking your expense</h2>
        {!isMobileDevice && 
          <div className="expense-interval-chart-wrapper" >
            <div><ExpenseChart intervalType="month" expenseObject={latestMonthExpense} /></div>
            <div><ExpenseChart intervalType="bi week" expenseObject={latestBiWeekExpense} /></div>
            <div><ExpenseChart intervalType="week" expenseObject={latestWeekExpense} /></div>
          </div>
        }

        <div className="interval-button-wrapper">
          <div className="interval-button-list">
            <button
              onClick={() => resetInterval("Month")}
              style={setButtonColor("Month")}>Month</button>
            <button 
              onClick={() => resetInterval("Bi_week")}
              style={setButtonColor("Bi_week")}>Bi Week</button>
            <button 
              onClick={() => resetInterval("Week")}
              style={setButtonColor("Week")}>Week</button>
          </div>
        </div>
        {isMobileDevice && 
          <div className="mobile-expense-interval-chart">
            <ExpenseChart
              intervalType={intervalType.toLowerCase()}
              expenseObject={latestIntervalExpense(intervalObject, intervalType)} />
          </div>
        }
        <div className="summary-info-wrapper">
          {/* the panels showing latest intervals of certaint type */}
          <div className="interval-panel-wrapper">
            <h3>5 Latest {
              intervalType === "Bi_week" ? (<span>Bi-Week</span>) : (<span>{intervalType}</span>)
            }</h3>

            <div className="super-interval-wrapper">
              { /* list all the intervals */
                latestIntervalArr.map((interval, index) =>
                  <div key={index} className="interval-wrapper"
                    onClick={() => onIntervalClick(index, selectedFirstDate, selectedLastDate)}
                    // overriding the color if this interval wrapper is clicked
                    style={{
                      color: index === clickedInterval && "rgba(0, 0, 0, 1)",
                      borderColor: index === clickedInterval && "rgba(0, 0, 0, 1)",
                    }}>
                    {reformatDate(interval.first_date, '-', '/')} to {reformatDate(interval.last_date, '-', '/')}
                  </div>
                )}
            </div>
          </div>

          <div className="summary-chart-wrapper">
            <div className="timeseries-chart">
              <TimeseriesChart intervalType={displayedType} timeseriesObject={intervalDailyExpense} />
            </div>
            <div className="change-chart">
              <ChangeChart intervalType={displayedType} changeObject={intervalChange} />
            </div>
            <div className="composition-chart" >
              <CompositionChart 
                intervalType={displayedType}
                compositionObject={intervalComposition} 
                handleCategoryClick={(e, chartRef) => 
                  onCategoryClick(e, chartRef, selectedFirstDate, selectedLastDate)} />
            </div>
            <p>Click on category to show the transactions during the period</p>
          </div>
        </div>

        {/* The transaction table */}
        <div className="summary-transaction-table">
          <TransactionTable
            listName={
              intervalTableObj.title + " Transactions From " +
              reformatDate(selectedFirstDate, '-', '/') + " To " + reformatDate(selectedLastDate, '-', '/')
            }
            transactionCount = {intervalTableObj.transactionCount} 
            transactionList={intervalTableObj.transactionList} />
        </div>
      </div>
    );
  }
  // if the data hasn't been loaded yet 
  else {
    return (<div>Is Loading...</div>);
  }
}

export default SummaryPage; 