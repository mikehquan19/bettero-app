import { useEffect, useState, MouseEvent, CSSProperties } from 'react';
import { getFullFinancialSummary, getInitialTransactions, getSummaryTransactions } from '@provider/api';
import handleError from '@provider/handleError';
import ChangeChart from '@Charts/ChangeChart';
import CompositionChart from '@Charts/CompositionChart';
import TimeseriesChart from '@Charts/TimeseriesChart';
import ExpenseChart from '@Charts/ExpenseChart';
import TransactionTable from '@components/TransactionTable/TransactionTable';
import { getElementAtEvent } from 'react-chartjs-2';
import { useMediaQuery } from '@uidotdev/usehooks';
import { reformatDate, latestIntervals, latestIntervalExpense, latestIntervalChart } from '@utils';
import { Interval, Transaction } from '@interface';
import './SummaryPage.scss';

interface SummaryPageProps {
  navbarWidth: number, 
  titleHeight: number
}

export default function SummaryPage({ navbarWidth, titleHeight }: SummaryPageProps) {
  const isMobileDevice = useMediaQuery("only screen and (max-width : 500px)");

  // Latest intervals and total expense of each interval
  const [intervalObject, setIntervalObject] = useState({ month: [], biWeek: [], week: [] });
  const [isLoading, setIsLoading] = useState<boolean>(true);

  const [intervalType, setIntervalType] = useState("month");
  const [clickedInterval, setClickedInterval] = useState(0);
  const [intervalTableObj, setIntervalTableObj] = useState({
    title: "All",
    transactionCount: 0,
    transactionList: [] as Transaction[],
  });

  /** 
   * Fetch the financial summary data 
   */
  function fetchFinancialSummary() {
    Promise.all([getFullFinancialSummary(), getInitialTransactions()])
      .then((response) => {
        setIntervalObject(response[0].data);
        setIntervalTableObj((previousData) => ({ ...previousData, ...response[1].data }));
        setIsLoading(false);
      })
      .catch((error) => handleError(error));
  }

  /** 
   * Fetch data about transactions during the period 
   */
  function fetchPeriodTransactions(): void {
    const latestIntervalArr = latestIntervals(intervalObject, intervalType);

    if (latestIntervalArr.length !== 0) {
      const firstDate = latestIntervalArr[clickedInterval].firstDate;
      const lastDate = latestIntervalArr[clickedInterval].lastDate;

      getSummaryTransactions("interval", undefined, firstDate, lastDate)
        .then((response) => setIntervalTableObj({ 
          title: "All", ...response.data
        }))
        .catch(error => handleError(error));
    }
  }

  /**
   * Fetch data about interval transactions based on click 
   */
  function onIntervalClick(intervalIndex: number, firstDate: string, lastDate: string): void {
    if (intervalIndex !== clickedInterval) {
      setClickedInterval(intervalIndex);
    } else {
      getSummaryTransactions("interval", undefined, firstDate, lastDate)
        .then((response) => setIntervalTableObj({ 
          title: "All", ...response.data 
        }))
        .catch((error) => handleError(error));
    }
  }

  /**
   * Fetch data of categorical transactions based on the click
   */
  function handleCategoryClick(
    e: MouseEvent<HTMLCanvasElement>, argChartRef: any, firstDate: string, lastDate: string
  ): void {
    const categoryList = ["Housing", "Automobile", "Medical", "Subscription", "Grocery", "Dining", "Shopping", "Gas", "Others"];
    const selectedElement = getElementAtEvent(argChartRef.current, e)[0];

    if (selectedElement) {
      const selectedCategory = categoryList[selectedElement.index];
      getSummaryTransactions("both", selectedCategory, firstDate, lastDate)
        .then((response) => setIntervalTableObj({ 
          title: selectedCategory, ...response.data 
        }))
        .catch((error) => handleError(error));
    }
  }

  /**
   * Handle the next click for the table 
   */
  function handleNextClick(pageIndex: number, firstDate: string, lastDate: string): void {
    const axiosResponse = intervalTableObj.title === 'All' ? 
      getSummaryTransactions("interval", undefined, firstDate, lastDate, pageIndex) : 
      getSummaryTransactions("both", intervalTableObj.title, firstDate, lastDate, pageIndex);
    
    axiosResponse
      .then((response) => setIntervalTableObj((previousData) => ({ 
        ...previousData, ...response.data
      })))
      .catch((error) => handleError(error));
  }

  /**
   * Reset the interval type, which in turn will reset the data 
   * @param {string} intervalType
   * @returns {void}
   */
  function resetInterval(intervalType: string): void {
    setIntervalType(intervalType);
    setClickedInterval(0);
  }

  /**
   * Set the button color 
   * @param {string} interval 
   * @returns {CSSProperties}
   */
  function setButtonColor(interval: string): CSSProperties {
    if (intervalType === interval) {
      return { backgroundColor: "rgba(0, 139, 139, 1)" } as CSSProperties;
    } else {
      return { backgroundColor: "rgba(0, 139, 139, 0)" } as CSSProperties;
    }
  }

  useEffect(() => fetchFinancialSummary(), []);
  // Only call after re-rendering triggered by interval type and clicked interval
  useEffect(() => { fetchPeriodTransactions() }, [intervalType, clickedInterval]);

  if (isLoading) return <div>Is Loading...</div>;

  // Type of interval to be displayed 
  const displayedType = intervalType === "biWeek" ? "bi-week" : intervalType.toLowerCase();

  // Update the data about total expense of 5 latest intervals 
  const latestMonthExpense = latestIntervalExpense(intervalObject, "month");
  const latestBiWeekExpense = latestIntervalExpense(intervalObject, "biWeek");
  const latestWeekExpense = latestIntervalExpense(intervalObject, "week");

  // Update the interval data based on the state  
  const latestIntervalArr = latestIntervals(intervalObject, intervalType);

  // Clicked first and last date
  const selectedFirstDate = latestIntervalArr[clickedInterval].firstDate;
  const selectedLastDate = latestIntervalArr[clickedInterval].lastDate;

  // Object for change and composition data of the clicked interval
  const intervalChartObject = latestIntervalChart(intervalObject, intervalType, selectedFirstDate);

  return (
    <div style={{ marginTop: `${titleHeight}px`, marginLeft: `${navbarWidth}px`, width: `calc(100% - ${navbarWidth}px)` }}>
      <h2 style={{ width: "85%", margin: "0 auto" }}>
        Detailed summary tracking your expense
      </h2>
      {!isMobileDevice &&
        /* The expense charts used for non-mobile screens */
        <div className="expense-interval-chart-wrapper" >
          <div>
            <ExpenseChart intervalType="month" expenseObject={latestMonthExpense} />
          </div>
          <div>
            <ExpenseChart intervalType="bi week" expenseObject={latestBiWeekExpense} />
          </div>
          <div>
            <ExpenseChart intervalType="week" expenseObject={latestWeekExpense} />
          </div>
        </div>
      }

      <div className="interval-button-wrapper">
        <div className="interval-button-list">
          <button onClick={() => resetInterval("month")} style={setButtonColor("month")}>Month</button>
          <button onClick={() => resetInterval("biWeek")} style={setButtonColor("biWeek")}>Bi Week</button>
          <button onClick={() => resetInterval("week")} style={setButtonColor("week")}>Week</button>
        </div>
      </div>
      {isMobileDevice &&
        /* The expense chart used for mobile screen, can change content dynamically */
        <div className="mobile-expense-interval-chart">
          <ExpenseChart
            intervalType={intervalType.toLowerCase()}
            expenseObject={latestIntervalExpense(intervalObject, intervalType)} />
        </div>
      }
      <div className="summary-info-wrapper">
        {/* the panels showing latest intervals of certaint type */}
        <div className="interval-panel-wrapper">
          <h3>
            5 Latest {intervalType === "biWeek" ? "Bi-Week" : intervalType}
          </h3>

          <div className="super-interval-wrapper">
            {
              /* list of all the intervals */
              latestIntervalArr.map((interval: Interval, index: number) =>
                <div key={index} className="interval-wrapper"
                  onClick={() => onIntervalClick(index, selectedFirstDate, selectedLastDate)}
                  style={{
                    // overriding the color if this interval wrapper is clicked
                    color: index === clickedInterval ? "rgba(0, 0, 0, 1)" : undefined,
                    borderColor: index === clickedInterval ? "rgba(0, 0, 0, 1)" : undefined,
                  }}
                >
                  {reformatDate(interval.firstDate, '-', '/')} to {reformatDate(interval.lastDate, '-', '/')}
                </div>
              )
            }
          </div>
        </div>

        <div className="summary-chart-wrapper">
          <div className="timeseries-chart">
            <TimeseriesChart
              contentType="expense"
              intervalType={displayedType} timeSeriesObject={intervalChartObject!.dailyExpense} />
          </div>
          <div className="change-chart">
            <ChangeChart intervalType={displayedType} changeObject={intervalChartObject!.changePercentage} />
          </div>
          <div className="composition-chart" >
            <CompositionChart
              intervalType={displayedType} chartType="regular"
              compositionObject={intervalChartObject!.compositionPercentage}
              onCategoryClick={(e, chartRef) => {
                handleCategoryClick(e, chartRef, selectedFirstDate, selectedLastDate);
              }}
            />
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
          transactionCount={intervalTableObj.transactionCount} 
          transactionList={intervalTableObj.transactionList}
          onChangeTransList={(pageIndex) => handleNextClick(pageIndex, selectedFirstDate, selectedLastDate)} />
      </div>
    </div>
  );
}