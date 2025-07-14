import { useState, useEffect } from 'react';
import CompositionChart from '../../Charts/CompositionChart';
import BillTable from '../../BillTable/BillTable';
import BudgetPlanForm from '../../Forms/AddForms/BudgetPlanForm/BudgetPlanForm';
import DeletePlanForm from '../../Forms/DeleteForms/DeletePlanForm';
import { expenseappClient } from '../../../provider/api';
import handleError from '../../../provider/handleError';
import { faPlusCircle, faCaretDown } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import './BudgetPage.css';

// the progress bar 
function ProgressBar({ categoryExpenseData }) {
  const [showCurrent, setShowCurrent] = useState(false);
  const [categoryInfo, setCategoryInfo] = useState({
    name: "",
    budget: 0,
    current: 0,
    percentage: 0,
  });

  useEffect(() => { setCategoryInfo({ ...categoryExpenseData }); }, [categoryExpenseData]);

  return (
    <div className="budget-bar">
      <label htmlFor="progress-part">
        <strong>{categoryInfo.name}:</strong>
      </label>
      <div className="number-part">
        <span className="display-budget">
          Budget: <strong>${categoryInfo.budget.toFixed(2)}</strong>
        </span>
        <span className="display-left">
          <strong>${(categoryInfo.budget - categoryInfo.current).toFixed(2)}</strong> left
        </span>
      </div>
      {categoryInfo.current >= categoryInfo.budget && (
        <small style={{ color: "red", display: "block", marginBottom: "10px", fontSize: "0.9rem", fontWeight: "500" }}>
          You have used up budget of this category. Be more cautious!
        </small>
      )}
      <div className="progress-bar">
        <div id="progress-part" className="filled-area" style={{ width: `${categoryInfo.percentage}%` }}>
        </div>
      </div>
      <span className="percentage-part" style={{left: `${categoryInfo.percentage - 4}%`}}
        onMouseOver={() => setShowCurrent(!showCurrent)}
        onMouseOut={() => setShowCurrent(!showCurrent)}>
        <strong>
          {showCurrent ?
            <span>${categoryInfo.current.toFixed(2)}</span> :
            <span>{categoryInfo.percentage.toFixed(2)}%</span>
          }
        </strong>
      </span>
    </div>
  );
}

// the panel showing the budget plan of the user 
function BudgetPlanPanel({
  intervalType, estimatedIncome, budgetPecentage, budgetChartData,
  actualChartData, progressData, handleChangePlanList = null }) {

  // whether the form is present or not 
  const [addFormPresent, setAddFormPresent] = useState(false);
  const [updateFormPresent, setUpdateFormPresent] = useState(false);
  const [deleteFormPresent, setDeleteFormPresent] = useState(false);

  // whether one button is disabled because the other one is 
  const [updateButtonDisabled, setUpdateButtonDisabled] = useState(false);
  const [deleteButtonDisabled, setDeleteButtonDisabled] = useState(false);

  // prepare the data that will be passed to the update form 
  let existingData;
  if (Object.keys(budgetChartData).length) {
    existingData = {
      interval_type: intervalType === "bi-week" ? "bi_week" : intervalType,
      recurring_income: estimatedIncome,
      portion_for_expense: budgetPecentage
    };
    Object.keys(budgetChartData).forEach((category) => {
      existingData[category] = budgetChartData[category];
    });
  }

  // everytime the user changes interval type, reset everything
  useEffect(() => {
    setAddFormPresent(false);
    setUpdateFormPresent(false);
    setUpdateButtonDisabled(false);
    setDeleteButtonDisabled(false);
  }, [intervalType]);

  // handle the disable of update or delete buttons 
  function handleDisable(type) {
    if (type == "UPDATE") {
      setUpdateFormPresent(!updateFormPresent);
      setDeleteButtonDisabled(!deleteButtonDisabled);
    } 
    else if (type == "DELETE") {
      setDeleteFormPresent(!deleteFormPresent);
      setUpdateButtonDisabled(!updateButtonDisabled);
    }
  }

  return (
    <>
      <h2 className="budget-plan-header">The budget plan for the {intervalType}:</h2>
      <div className="budget-plan-panel">
        {Object.keys(budgetChartData).length === 0 ? (
          <>
            <h3 className='none-message'>
              You currently don't have a plan for the {intervalType}
            </h3>
            <button className="add-plan-button" onClick={() => setAddFormPresent(!addFormPresent)}>
              <FontAwesomeIcon icon={faPlusCircle} style={{ display: "block", margin: "0 auto 5px auto" }} />
              Add Plan
            </button>

            {addFormPresent && <BudgetPlanForm
              type="ADD"
              intervalType={intervalType === "bi-week" ? "bi_week" : intervalType}
              handleHideForm={() => setAddFormPresent(!addFormPresent)}
              handleChangePlanList={handleChangePlanList} />}
          </>
        ) : (
          <><div className="income-and-budget">
            <h3>Your estimated income this {intervalType}: <span style={{fontSize: "25px"}}>${estimatedIncome}</span></h3>
            <h3>You put <span style={{fontSize: "25px"}}>{budgetPecentage}%</span> of your income on expense</h3>
          </div>

          <div className="composition-charts-part">
            <div>
              <CompositionChart chartType="budget" chartWidth="410"
                intervalType={intervalType} compositionObject={budgetChartData} />
            </div>
            <div>
              <CompositionChart chartType="actual" chartWidth="410"
                intervalType={intervalType} compositionObject={actualChartData} />
            </div>
          </div>

          <div className="progress-bars-part">
            {Object.keys(progressData).map(category =>
              <ProgressBar key={category} categoryExpenseData={{
                name: category,
                budget: progressData[category].budget,
                current: progressData[category].current,
                percentage: progressData[category].percentage,
              }} />
            )}
          </div>

          <div className="budget-button-list">
            <button disabled={updateButtonDisabled} onClick={() => handleDisable("UPDATE")}>Update</button>
            <button disabled={deleteButtonDisabled} onClick={() => handleDisable("DELETE")}>Delete</button>
          </div>

          {updateFormPresent && (<BudgetPlanForm
            type="UPDATE"
            existingData={existingData}
            handleHideForm={() => handleDisable("UPDATE")} handleChangePlanList={handleChangePlanList} />)
          }

          {deleteFormPresent && (<DeletePlanForm
            intervalType={intervalType === "bi-week" ? "bi_week" : intervalType}
            handleHideForm={() => handleDisable("DELETE")}
            handleChangePlanList={handleChangePlanList} />)
          }</>
        )}
      </div>
    </>
  );
}

// the entire budget page 
const BudgetPage = ({ navbarWidth, titleHeight }) => {
  const initialData = {
    income: 0,
    expense_portion: 0,
    composition: {
      goal: {},
      actual: {}
    },
    progress: {}
  }

  // fetch data from the api 
  // initial data just for the sake of initial rendering
  const [budgetData, setBudgetData] = useState({ month: initialData, bi_week: initialData, week: initialData });
  const [overdueMessageList, setOverdueMessageList] = useState([]);
  const [overdueMessageListPresent, setOverdueMessageListPresent] = useState(false);
  const [caretIconStyle, setCaretIconStyle] = useState({
    marginLeft: "10px",
    fontSize: "22px",
    transition: "transform 0.5s linear",
  });

  // state of the form 
  const [intervalType, setIntervalType] = useState("month");

  // update chart data 
  let budgetChartData = {};
  let actualChartData = {};
  let progressData = {};
  let estimatedIncome = 0;
  let budgetPecentage = 0;

  if (Object.keys(budgetData[intervalType]).length) {
    budgetChartData = budgetData[intervalType].composition.goal;
    actualChartData = budgetData[intervalType].composition.actual;
    progressData = budgetData[intervalType].progress;
    estimatedIncome = budgetData[intervalType].income;
    budgetPecentage = budgetData[intervalType].expense_portion
  }

  // type to be display
  const displayedType = intervalType === "bi_week" ? "bi-week" : intervalType;

  function setButtonColor(interval) {
    if (intervalType === interval) {
      return { backgroundColor: "rgba(0, 139, 139, 1)" };
    } else {
      return { backgroundColor: "rgba(0, 139, 139, 0)" };
    }
  }

  function onCarretClick() {
    if (!overdueMessageListPresent) {
      setOverdueMessageListPresent(true);
      setCaretIconStyle({
        marginLeft: "10px",
        fontSize: "20px",
        transform: "rotate(180deg)",
        transition: "transform 0.5s linear",
      });
    } 
    else {
      setOverdueMessageListPresent(false);
      setCaretIconStyle({
        marginLeft: "10px",
        fontSize: "20px",
        transition: "transform 0.5s linear",
      });
    }
  }

  useEffect(() => {
    Promise.all([
      expenseappClient.get('/budget'), // the list of budgets of the user 
      expenseappClient.get(`/overdue_message`) // the list of overdue bill message
    ])
      .then((response) => {
        setBudgetData(response[0].data);
        setOverdueMessageList(response[1].data);
      })
      .catch((error) => handleError(error));
  }, []);

  return (
    <div className="budget-page" style={{
      marginTop: `${titleHeight}px`,
      marginLeft: `${navbarWidth}px`,
      width: `calc(100% - ${navbarWidth}px)`,
    }}>
      <h2 style={{ width: "80%", margin: "0 auto" }}>List of your monthly bills and budget plans</h2>

      {overdueMessageList.length !== 0 && 
        <div className="overdue-message-list">
          <h3 onClick={onCarretClick}>
            **You have {overdueMessageList.length} overdue bills
            <FontAwesomeIcon icon={faCaretDown} style={caretIconStyle} />
          </h3>
          {overdueMessageListPresent && 
            <div className="overdue-message-wrapper">
              {overdueMessageList.map((message, index) =>
                <div className="overdue-message" key={index}>
                  <p><strong>{index + 1}/</strong> {message.bill_description}: ${message.bill_amount}, due on <strong>{message.bill_due_date}</strong></p>
                </div>
              )}
            </div>
          }
        </div>
      }

      {/* The table of list of bills  */}
      <BillTable />

      {/* The panel consisting of buttons of interval type */}
      <div className="interval-button-wrapper">
        <div className="interval-button-list">
          <button onClick={() => setIntervalType("month")} style={setButtonColor("month")}>Month</button>
          <button onClick={() => setIntervalType("bi_week")} style={setButtonColor("bi_week")}>Bi Week</button>
          <button onClick={() => setIntervalType("week")} style={setButtonColor("week")}>Week</button>
        </div>
      </div>

      {/* The panel showing the budget plan */}
      <BudgetPlanPanel
        intervalType={displayedType}
        estimatedIncome={estimatedIncome} budgetPecentage={budgetPecentage}
        budgetChartData={budgetChartData} actualChartData={actualChartData}
        progressData={progressData} handleChangePlanList={(newBudgetData) => setBudgetData(newBudgetData)} />
    </div>
  );
}

export default BudgetPage;