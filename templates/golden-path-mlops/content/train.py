import mlflow
import mlflow.sklearn
from sklearn.ensemble import RandomForestClassifier
from sklearn.model_selection import train_test_split
from sklearn.metrics import accuracy_score
import numpy as np

MLFLOW_TRACKING_URI = "http://mlflow.mlops.svc.cluster.local:5000"
mlflow.set_tracking_uri(MLFLOW_TRACKING_URI)
mlflow.set_experiment("${{ values.name }}")

def train():
    X = np.random.rand(200, 4)
    y = (X[:, 0] + X[:, 1] > 1).astype(int)
    X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2)

    with mlflow.start_run():
        n_estimators = 100
        model = RandomForestClassifier(n_estimators=n_estimators)
        model.fit(X_train, y_train)
        acc = accuracy_score(y_test, model.predict(X_test))

        mlflow.log_param("n_estimators", n_estimators)
        mlflow.log_metric("accuracy", acc)
        mlflow.sklearn.log_model(model, "${{ values.name }}")

        print(f"[${{ values.name }}] accuracy={acc:.4f}")

if __name__ == "__main__":
    train()
