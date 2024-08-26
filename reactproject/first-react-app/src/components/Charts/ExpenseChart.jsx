import {
    Chart as ChartJS,
    CategoryScale,
    LinearScale,
    BarElement,
    Title,
    Tooltip,
    Legend,
} from 'chart.js';
import { Bar } from 'react-chartjs-2';
import 'chartjs-adapter-date-fns';

// register the tool to be able to to use it in the hand-shakeable way 
ChartJS.register(CategoryScale, LinearScale, BarElement, 
    Title, Tooltip, Legend);

// chart indicating the change percentage from the previous month 
const ExpenseChart = ({intervalType = "month" , expenseObject}) => {

    // list of categories of expense for labels 
    const dateLabels = Object.keys(expenseObject).reverse();
    const expenseData = Object.values(expenseObject).reverse(); 
    

    // data of chart 
    const chartData = {
        labels: dateLabels,
        datasets: [
            {
                label: "the expense of the given " + intervalType,
                data: expenseData,
                backgroundColor: "rgba(0, 0, 255, 0.5)",
                // the color of the bar when hovered over
                hoverBackgroundColor: "rgb(255, 0, 0, 0.5)",
                hoverBorderColor: "black",
                hoverBorderWidth: 2,
            },
        ],
    }; 

    // options of the chart 
    const chartOptions = {
        // this is to gain control of the width and height of the canvas
        maintainAspectRatio: false,
        responsive: true,
        // make the bar chart horizontal  
        scales: {
            x: {
                grid: {
                    color: "skyblue",
                }
            },
        },
        plugins: {
            // set the title for the chart 
            title: {
                display: true,
                color: 'black',
                text: 'Total expense ' + intervalType,
                font: {
                    size: 16,
                }
            }, 
        },
    };

    // render the bar chart 
    return <Bar data={chartData} options={chartOptions} />;
}

export default ExpenseChart