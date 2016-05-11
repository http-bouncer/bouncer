/**
 * Created by tugceakin on 4/8/16.
 */

bouncerApp.controller('BenchmarkingController', function($scope, $interval, $http, $parse, benchmarking, configuration) {
    $scope.graphOff = true;
    $scope.benchmarkCompleted = false;
    $scope.benchmarkInput = "";
    $scope.stats = {};
    $scope.statsShown = false;
    $scope.selectedConfig = {};
    $scope.backendServers = [];

    benchmarking.setSocketConnection;
    benchmarking.resetGraph($scope);
    benchmarking.onGraphLineClick($scope);

    configuration.getAllConfigs($scope);

    $scope.reqPerSeqOptions = { animation: false }
    $scope.responseTimeOptions = { animation: false }
    $scope.status200Options = { animation: false }
    $scope.status404Options = { animation: false }

    //$scope.startBenchmarking = function(){
        benchmarking.resetGraph($scope);
        benchmarking.updateGraph($scope);
    //}

    $scope.closeConnection = function(config){
        benchmarking.closeConnection(this.selectedConfig);
        $scope.graphOff = true;
    }

    $scope.getBenchmarkingStats = function(){
        if($scope.statsShown == false){
            $scope.statsShown = true;
        }else{
            $scope.statsShown = false;
        }
    }

    $scope.configurationSelected = function(config){
        console.log(this.selectedConfig);
        if(this.selectedConfig != null){
            //$scope.currentConfig = config;
            benchmarking.resetGraph($scope);
            benchmarking.updateGraph($scope, this.selectedConfig);
            //benchmarking.sendConnection($scope, this.selectedConfig);
        }

    }


});