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
import './SummaryPage.css';

// reformat the date for displaying on the transactions table and interval panel 
const reformatDate = (argDate) => {
  const dateArray = argDate.split("-");
  const year = dateArray[0];
  const month = dateArray[1];
  const day = dateArray[2];
  return (month + "/" + day + "/" + year);
}

// return the list of latest intervals of the predefined interval type
const latestIntervals = (intervalContent, intervalType) => {
  let latestIntervalArr = [];
  let latestIntervalData = intervalContent[intervalType.toLowerCase()];
  for (let i = 0; i < latestIntervalData.length; i++) {
    // create the interval 
    let interval = {};
    interval["first_date"] = latestIntervalData[i].first_date;
    interval["last_date"] = latestIntervalData[i].last_date;
    // add interval to the array of intervals
    latestIntervalArr.push(interval);
  }
  return latestIntervalArr;
}

// return the list of expense of latest intervals of predefined interval type
const latestIntervalExpense = (intervalContent, intervalType) => {
  let intervalExpenseObj = {};
  let intervalExpenseData = intervalContent[intervalType.toLowerCase()];
  for (let i = 0; i < intervalExpenseData.length; i++) {
    let firstDate = intervalExpenseData[i].first_date;
    let totalExpense = intervalExpenseData[i].total_expense;

    // add the total expense of the interval to the object 
    intervalExpenseObj[firstDate] = totalExpense;
  }
  return intervalExpenseObj;
}


// return the object mapping the specific interval
// to the expense change and expense composition 
const latestIntervalChart = (intervalContent, intervalType, firstDate) => {
  let intervalObject = {};
  let intervalChartData = intervalContent[intervalType.toLowerCase()];

  for (let i = 0; i < intervalChartData.length; i++) {
    const pickedItem = intervalChartData[i];
    // if the first date of the interval matches 
    if (pickedItem.first_date === firstDate) {
      // add the daily expense, expense change, and composition to the object
      intervalObject["daily_expense"] = pickedItem.daily_expense;
      intervalObject["change_percentage"] = pickedItem.expense_change;
      intervalObject["composition_percentage"] = pickedItem.expense_composition;
    }
  }
  return intervalObject;
}

// the summary page component 
const SummaryPage = ({ navbarWidth, titleHeight }) => {
  const isMobileDevice = useMediaQuery("only screen and (max-width : 500px)");

  // fetch data about the latest intervals and total expense of each interval
  const [intervalContent, setIntervalContent] = useState({"month": [], "bi-week": [], "week": []});
  const [isLoading, setIsLoading] = useState(true);

  const [intervalType, setIntervalType] = useState("Month");
  const [clickedInterval, setClickedInterval] = useState(0);
  const [transTableContent, setTransTableContent] = useState({
    title: "All",
    transactionList: [],
  });

  // call the API to fetch the data about financial summary from backend 
  useEffect(() => {
    expenseappClient.get("/full_summary")
      .then((response) => {
        setIntervalContent(response.data.latest_interval_expense);
        setTransTableContent((previousData) => ({
          ...previousData,
          transactionList: response.data.initial_transaction_data,
        }))
        setIsLoading(false);
      })
      .catch((error) => handleError(error));
  }, []);

  // call the API to fetch data about transactions during the period 
  useEffect(() => {
    const latestIntervalArr = latestIntervals(intervalContent, intervalType);
    // only call after the re-rendering triggered by the change of intervalType and clickedInterval
    if (latestIntervalArr.length !== 0) {
      const firstDate = latestIntervalArr[clickedInterval].first_date;
      const lastDate = latestIntervalArr[clickedInterval].last_date;

      // for checking 
      expenseappClient.get(`/transactions/interval?first_date=${firstDate}&last_date=${lastDate}`)
        .then((response) => {
          setTransTableContent({
            title: "All",
            transactionList: response.data,
          });
        })
        .catch(error => handleError(error));
    }
  }, [intervalType, clickedInterval]);


  // if the data has been loaded entirely 
  if (!isLoading) {
    // update the data for the total expense of 5 latest intervals 
    const latestMonthExpense = latestIntervalExpense(intervalContent, "Month");
    const latestBiWeekExpense = latestIntervalExpense(intervalContent, "Bi_week");
    const latestWeekExpense = latestIntervalExpense(intervalContent, "Week");

    // update the interval data based on the state 
    // update the list of intervals 
    const latestIntervalArr = latestIntervals(intervalContent, intervalType)

    // the first date and last date of the clicked interval 
    const clickedFirstDate = latestIntervalArr[clickedInterval].first_date;
    const clickedLastDate = latestIntervalArr[clickedInterval].last_date;

    // object containing change and composition data of the clicked interval
    const intervalChartObject = latestIntervalChart(intervalContent, intervalType, clickedFirstDate);

    // update the data for daily expense, change, and composition chart of that interval 
    const intervalDailyExpense = intervalChartObject.daily_expense;
    const intervalChange = intervalChartObject.change_percentage;
    const intervalComposition = intervalChartObject.composition_percentage;

    // type of interval to be displayed 
    const displayedType = intervalType === "Bi_week" ? "bi-week" : intervalType.toLowerCase();

    // handle category click (same function)
    const categoryList = ["Housing", "Automobile", "Medical", "Subscription", "Grocery", "Dining", "Shopping", "Gas", "Others"];
    const handleCategoryClick = (e, argChartRef) => {
      const selectedElement = getElementAtEvent(argChartRef.current, e)[0]
      // if the element is not undefined 
      if (selectedElement) {
        const newCategory = categoryList[selectedElement.index];
        expenseappClient.get(`/transactions/both?category=${newCategory}&first_date=${clickedFirstDate}&last_date=${clickedLastDate}`)
          .then((response) => {
            setTransTableContent({
              title: newCategory,
              transactionList: response.data,
            });
          })
          .catch((error) => handleError(error));
      }
    }

    const setButtonColor = (interval) => {
      if (intervalType === interval) {
        return {backgroundColor: "rgba(0, 139, 139, 1)"};
      } 
      return {backgroundColor: "rgba(0, 139, 139, 0)"};
    }

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
        {!isMobileDevice && (
          <div className="expense-interval-chart-wrapper" >
            <div><ExpenseChart intervalType="month" expenseObject={latestMonthExpense} /></div>
            <div><ExpenseChart intervalType="bi week" expenseObject={latestBiWeekExpense} /></div>
            <div><ExpenseChart intervalType="week" expenseObject={latestWeekExpense} /></div>
          </div>
        )}

        <div className="interval-button-wrapper">
          <div className="interval-button-list">
            <button
              onClick={() => {
                setIntervalType("Month");
                setClickedInterval(0); // reset the interval panel 
              }}
              style={setButtonColor("Month")}>Month</button>
            <button
              onClick={() => {
                setIntervalType("Bi_week");
                setClickedInterval(0); // reset the interval panel      
              }}
              style={setButtonColor("Bi_week")}>Bi Week</button>
            <button
              onClick={() => {
                setIntervalType("Week");
                setClickedInterval(0); // reset the interval panel 
              }}
              style={setButtonColor("Week")}>Week</button>
          </div>
        </div>
        {isMobileDevice && (
          <div className="mobile-expense-interval-chart">
            <ExpenseChart
              intervalType={intervalType.toLowerCase()}
              expenseObject={latestIntervalExpense(intervalContent, intervalType)} />
          </div>
        )}
        <div className="summary-info-wrapper">
          {/* the panels showing latest intervals of certaint type */}
          <div className="interval-panel-wrapper">
            <h3>5 Latest {intervalType === "Bi_week" ? (<span>Bi-Week</span>) : (<span>{intervalType}</span>)}</h3>

            <div className="super-interval-wrapper">
              { /* list all the intervals */
                latestIntervalArr.map((interval, index) =>
                  <div key={index} className="interval-wrapper"
                    onClick={() => {
                      if (index !== clickedInterval) {
                        setClickedInterval(index);
                      } 
                      else {
                        expenseappClient.get(`/transactions/interval?first_date=${clickedFirstDate}&last_date=${clickedLastDate}`)
                          .then((response) => {
                            setTransTableContent({
                              title: "All",
                              transactionList: response.data,
                            });
                          })
                          .catch((error) => handleError(error));
                      }
                    }}
                    // overriding the color if this interval wrapper is clicked
                    style={{
                      color: index === clickedInterval && "rgba(0, 0, 0, 1)",
                      borderColor: index === clickedInterval && "rgba(0, 0, 0, 1)",
                    }}>
                    {reformatDate(interval.first_date)} to {reformatDate(interval.last_date)}
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
              <CompositionChart intervalType={displayedType} compositionObject={intervalComposition}
                handleCategoryClick={handleCategoryClick} />
            </div>
            <p>Click on category to show the transactions during the period</p>
          </div>
        </div>

        {/* The transaction table */}
        <div className="summary-transaction-table">
          <TransactionTable
            listName={
              transTableContent.title + " Transactions From " +
              reformatDate(clickedFirstDate) +
              " To " + reformatDate(clickedLastDate)
            }
            transactionList={transTableContent.transactionList} />
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