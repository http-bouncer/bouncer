bouncerApp.directive("addbackendserverinput", function(){
	return {
		restrict: "E",
		template: '<button addinput id="add-input-button" class="btn btn-success btn-sm backendServerButtons">' +
           		 '<span class="glyphicon glyphicon-plus" aria-hidden="true"></span></button>' +
           		 '<button removeinput class="btn btn-danger btn-sm backendServerButtons" id="remove-input-button">' +
           		 '<span class="glyphicon glyphicon-minus" aria-hidden="true"></span></button>'
	}
});

bouncerApp.directive("addinput", function($compile){
	return function(scope, element, attrs){
		element.bind("click", function(){
			scope.backendServerCount++;
			angular.element(document.getElementById('backend-servers')).append(
				$compile("<input class='form-control backend-server-input backendServerInput' name= 'backendServer' ng-model='config.backendServer"+scope.backendServerCount+"' "
					+ "id='backendServerInput" + scope.backendServerCount +"' placeholder='Enter backend server name...'>")(scope)
				);
		});
	};
});

bouncerApp.directive("removeinput", function($compile){
	return function(scope, element, attrs){
		element.bind("click", function(){
			if(scope.backendServerCount > 1){
				angular.element(document.getElementById('backendServerInput' + scope.backendServerCount)).remove();
				scope.backendServerCount--;				
			}
		});
	};
});