from datetime import datetime, timedelta
from airflow import DAG
from airflow.operators.python import PythonOperator

default_args = {
    'owner': '${{ values.owner }}',
    'retries': 1,
    'retry_delay': timedelta(minutes=5),
}

def extract():
    print("[${{ values.name }}] Extract — lecture source données")

def transform():
    print("[${{ values.name }}] Transform — nettoyage et enrichissement")

def quality_check():
    print("[${{ values.name }}] Quality — validation Great Expectations")

def load():
    print("[${{ values.name }}] Load — écriture destination")

with DAG(
    dag_id='${{ values.name }}',
    description='${{ values.description }}',
    default_args=default_args,
    start_date=datetime(2026, 1, 1),
    schedule_interval='@daily',
    catchup=False,
    tags=['dxp', 'dataops'],
) as dag:
    t1 = PythonOperator(task_id='extract',       python_callable=extract)
    t2 = PythonOperator(task_id='transform',     python_callable=transform)
    t3 = PythonOperator(task_id='quality_check', python_callable=quality_check)
    t4 = PythonOperator(task_id='load',          python_callable=load)
    t1 >> t2 >> t3 >> t4
