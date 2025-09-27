from django.contrib import admin 
from django.urls import path, include 
from django.http import JsonResponse

def health_check(request):
    """ Used for EB healthcheck """
    return JsonResponse({"status": "ok"})

urlpatterns = [
    path('health/', health_check),
    path('expenseapp/', include("expenseapp.urls")),
    path('admin/', admin.site.urls),
]
