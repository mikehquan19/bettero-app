import { useState, useEffect } from "react";
import { useForm } from "react-hook-form";
import { faX } from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { expenseappClient } from "@provider/api";
import handleError from "@provider/handleError";
import "./BudgetPlanForm.css";
import { BudgetPlan } from "@interface";

interface BudgetPlanFormProps {
  type: string, 
  intervalType?: string, 
  currentData?: any, 
  onChangeBudgetPlanList?: (data: BudgetPlan) => void, 
  onHide: () => void,
}

// the form to add or update the budget plan 
function BudgetPlanForm({
  type = "ADD", intervalType, currentData, onChangeBudgetPlanList, onHide 
}: BudgetPlanFormProps) {

  const categoryArr = ["Total", "Housing", "Automobile", "Medical", "Subscription", "Grocery", "Dining", "Shopping", "Gas", "Others"];

  // necessary components of useForm() hook 
  const { register, reset, handleSubmit, formState: { errors } } = useForm();
  const [percentageNotAddUp, setPrecentageNotAddUp] = useState(false);

  // set the initial value for the form 
  useEffect(() => {
    if (!!currentData) {
      reset({ ...currentData });
    } else {
      reset((previousData) => ({
        ...previousData,
        interval_type: intervalType,
      }));
    }
  }, []);

  // handle the submission of the form 
  function budgetPlanFromServer(type: string, intervalType: string, data: any) {
    if (type === "ADD") {
      return expenseappClient.post("/budget", data);
    } else if (type === "UPDATE") {
      return expenseappClient.put(`budget/${intervalType}`, data);
    } else {
      throw Error("Invalid type of form!");
    }
  }

  function handleBudgetFormSubsmit (data: any) {

    const submittedData = {
      interval_type: data.interval_type,
      recurring_income: data.recurring_income, 
      portion_for_expense: data.portion_for_expense,
      category_portion: {} as Record<string, string>
    };

    let sum = 0; // used to validate
    categoryArr.slice(1, categoryArr.length).forEach((category) => {
      submittedData.category_portion[category] = data[category]
      sum += Number(data[category])
    })

    // validate if the all categories add up 100%
    if (sum !== 100) {
      setPrecentageNotAddUp(true);
    }
    else {
      setPrecentageNotAddUp(false);
      // call the correct API method based on the form's type
      const intervalType = !currentData ? null : currentData.interval_type;
      budgetPlanFromServer(type, intervalType, submittedData)
        .then((response) => !!onChangeBudgetPlanList ? onChangeBudgetPlanList(response.data) : null)
        .catch((error) => handleError(error));

      onHide();
    }
  }

  // validation on the form 
  const registerOptions = {
    recurring_income: {
      required: "*Income is required",
      min: {
        value: 0.01,
        message: "*Income must be greater than 0",
      },
      maxLength: {
        value: 12,
        message: "*Income only has at most 12 digits",
      }
    },
  } as Record<string, any>;

  categoryArr.forEach(category => {
    const field = category === 'Total' ? 'portion_for_expense' : category;
    registerOptions[field] = {
      required: `*${category} required`,
      min: {
        value: 0,
        message: "*Percentage must be greater than or equal 0",
      },
      max: {
        value: 100,
        message: "*Percentage must be less than or equal 100",
      }
    }
  })

  return (
    <>
      <div className="budget-plan-form-wrapper">
        <h3 style={{ marginTop: "0", textAlign: "center" }}>{type} BUDGET PLAN</h3>
        <form id="budget-plan-form" onSubmit={handleSubmit(handleBudgetFormSubsmit)}>

          <div className="category-field">
            <div className="form-field">
              <label htmlFor="estimated-income">Income ($):</label>
              <small className="error-message">
                {errors.recurring_income && errors.recurring_income.message as string}
              </small>
              <input type="number" id="estimated-income" placeholder="Estimated income"
                {...register("recurring_income", registerOptions.recurring_income)} />
            </div>

            <div className="form-field">
              <label htmlFor="total-budget">Expense portion (%):</label>
              <small className="error-message">
                {errors.portion_for_expense && errors.portion_for_expense.message as string}
              </small>
              <input type="number" id="total-budget" placeholder="Total budget"
                {...register("portion_for_expense", registerOptions.portion_for_expense)} />
            </div>
          </div>

          {/* some instructions */}
          <h4 style={{ textAlign: "center" }}>Enter the portion of total budget for each category</h4>
          <p className="error-message" style={{ textAlign: "center" }}>
            {percentageNotAddUp && "*All categories don't add up to 100%"}
          </p>

          <div className="category-field">
            { /* Each field representing one category budget */
              categoryArr.slice(1, categoryArr.length).map((category, index) => 
                <div key={index} className="form-field">
                  <label htmlFor={category}>{category}:</label>
                  <small className="error-message">
                    {errors[category] && errors[category].message as string}
                  </small>
                  <input type="number" id={category} placeholder={category}
                    {...register(category, registerOptions[category])} />
                </div>
              )
            }
          </div>
        </form>
        {/* The buttons for form */}
        <div className="plan-form-buttons">
          <button form="budget-plan-form" type="submit">Submit</button>
          <button type="button" onClick={onHide}>Close</button>
        </div>
        <button className="x-button" onClick={onHide}>
          <FontAwesomeIcon icon={faX} />
        </button>
      </div>
      <div className="overlay"></div>
    </>
  );
}

export default BudgetPlanForm; 