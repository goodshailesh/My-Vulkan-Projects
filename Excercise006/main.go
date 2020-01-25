package main

import (
	"fmt"

	"github.com/vulkan-go/glfw/v3.3/glfw"
	vk "github.com/vulkan-go/vulkan"
)

func main() {
	glfw.Init()
	vk.SetGetInstanceProcAddr(glfw.GetVulkanGetInstanceProcAddress())
	vk.Init()

	var appInfo *vk.ApplicationInfo = &vk.ApplicationInfo{
		SType:              vk.StructureTypeApplicationInfo,
		PNext:              nil,
		PApplicationName:   "myVulkan Application\x00",
		ApiVersion:         vk.MakeVersion(1, 0, 0), // Throws 'vulkan error: incompatible driver' error with incorrect version number
		ApplicationVersion: vk.MakeVersion(1, 0, 0),
		PEngineName:        "My Game Engine\x00",
		EngineVersion:      vk.MakeVersion(0, 1, 0),
	}
	// Resources
	var instance vk.Instance
	var physicalDevices []vk.PhysicalDevice
	var physicalDeviceProperties vk.PhysicalDeviceProperties
	var physicalDeviceFeatures vk.PhysicalDeviceFeatures
	var memoryProperties vk.PhysicalDeviceMemoryProperties
	var pQueueFamilyProperties []vk.QueueFamilyProperties
	//var logicalDevice vk.Device

	//Create Instance
	var layers = []string{"VK_LAYER_KHRONOS_validation\x00"}
	// For Windows Only
	var extensions = []string{"VK_KHR_surface\x00", "VK_KHR_win32_surface\x00"}
	// For Linux
	//https://software.intel.com/en-us/articles/api-without-secrets-introduction-to-vulkan-part-2
	//var extensions = []string{"VK_KHR_surface\x00", "VK_KHR_xcb_surface\x00"}
	var instanceInfo = vk.InstanceCreateInfo{
		SType: vk.StructureTypeInstanceCreateInfo,
		PNext: nil,
		//Flags:                   nil,
		PApplicationInfo:        appInfo,
		EnabledLayerCount:       uint32(len(layers)),
		PpEnabledLayerNames:     layers,
		EnabledExtensionCount:   uint32(len(extensions)),
		PpEnabledExtensionNames: extensions,
	}
	result := vk.CreateInstance(&instanceInfo, nil, &instance)
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Error creating instance: %v", result))
	}

	physicalDevices = getPhysicalDevices(instance)
	physicalDeviceProperties = getPhysicalDeviceProperties(physicalDevices[0])
	physicalDeviceFeatures = getPhysicalDeviceFeatures(physicalDevices[0])
	memoryProperties = getPhysicalDeviceMemoryProperties(physicalDevices[0])
	pQueueFamilyProperties = getPhysicalDeviceQueueFamilyProperties(physicalDevices[0])

	// Get the memory properties of the physical device.
	vk.GetPhysicalDeviceMemoryProperties(physicalDevices[0], &memoryProperties)

	// Verbose - Please don't remove, igrnoe
	physicalDeviceProperties.Deref()
	fmt.Println(physicalDevices)
	fmt.Println(vk.ToString(physicalDeviceProperties.DeviceName[:]))
	fmt.Println(physicalDeviceFeatures)
	fmt.Println(memoryProperties)
	fmt.Println(pQueueFamilyProperties)
	uitlPrintDeviceQueueFamilyProperties(pQueueFamilyProperties)
}

// func createDevice(physicalDevice vk.PhysicalDevice) *vk.Device {
// 	var logicalDevicePtr *vk.Device

// 	result := vk.CreateDevice(physicalDevice, pCreateInfo*DeviceCreateInfo, nil, logicalDevicePtr)
// 	if result != vk.Success {
// 		fmt.Println(fmt.Errorf("Error getting physical device count: %v", result))
// 		return nil
// 	}
// }

func uitlPrintDeviceQueueFamilyProperties(qfs []vk.QueueFamilyProperties) {
	fmt.Println("Print device queue properties..............")
	fmt.Printf("\tFound %v families\n", len(qfs))
	for idx, qf := range qfs {
		qf.Deref()
		fmt.Printf("\tFamily %v\n", idx)
		flagBits := vk.QueueFlagBits(qf.QueueFlags)
		if flagBits&vk.QueueGraphicsBit != 0x00000000 {
			fmt.Printf("\t\tVK_QUEUE_GRAPHICS_BIT \t\t[vk.QueueGraphicsBit] \t\t[0x00000001\\1]\n")
		}
		if flagBits&vk.QueueComputeBit != 0x00000000 {
			fmt.Printf("\t\tVK_QUEUE_COMPUTE_BIT \t\t[vk.QueueComputeBit] \t\t[0x00000002\\2]\n")
		}
		if flagBits&vk.QueueTransferBit != 0x00000000 {
			fmt.Printf("\t\tVK_QUEUE_TRANSFER_BIT \t\t[vk.QueueTransferBit] \t\t[0x00000004\\4]\n")
		}
		if flagBits&vk.QueueSparseBindingBit != 0x00000000 {
			fmt.Printf("\t\tVK_QUEUE_SPARSE_BINDING_BIT \t[vk.QueueSparseBindingBit] \t[0x00000008\\8]\n")
		}
		if flagBits&vk.QueueProtectedBit != 0x00000000 {
			fmt.Printf("\t\tVK_QUEUE_PROTECTED_BIT \t\t[vk.QueueProtectedBit] \t\t[0x00000010\\16]\n")
		}
		if flagBits&vk.QueueFlagBitsMaxEnum != 0x00000000 {
			fmt.Printf("\t\tVK_QUEUE_FLAG_BITS_MAX_ENUM \t[QueueFlagBitsMaxEnumt] \t[0x7FFFFFFF\\2147483647]\n")
		}
	}
}

func getPhysicalDeviceQueueFamilyProperties(physicalDevice vk.PhysicalDevice) []vk.QueueFamilyProperties {
	var pQueueFamilyPropertyCount uint32
	vk.GetPhysicalDeviceQueueFamilyProperties(physicalDevice, &pQueueFamilyPropertyCount, nil)
	var pQueueFamilyProperties []vk.QueueFamilyProperties
	pQueueFamilyProperties = make([]vk.QueueFamilyProperties, pQueueFamilyPropertyCount)
	vk.GetPhysicalDeviceQueueFamilyProperties(physicalDevice, &pQueueFamilyPropertyCount, pQueueFamilyProperties)
	return pQueueFamilyProperties
}

func getPhysicalDeviceMemoryProperties(physicalDevice vk.PhysicalDevice) vk.PhysicalDeviceMemoryProperties {
	var pMemoryProperties vk.PhysicalDeviceMemoryProperties
	vk.GetPhysicalDeviceMemoryProperties(physicalDevice, &pMemoryProperties)
	return pMemoryProperties
}

func getPhysicalDeviceFeatures(physicalDevice vk.PhysicalDevice) vk.PhysicalDeviceFeatures {
	var pFeatures vk.PhysicalDeviceFeatures
	vk.GetPhysicalDeviceFeatures(physicalDevice, &pFeatures)
	return pFeatures
}

func getPhysicalDeviceProperties(physicalDevice vk.PhysicalDevice) vk.PhysicalDeviceProperties {
	var pProperties vk.PhysicalDeviceProperties
	vk.GetPhysicalDeviceProperties(physicalDevice, &pProperties)
	return pProperties
}

func getPhysicalDevices(instance vk.Instance) []vk.PhysicalDevice {
	fmt.Println("[PhysicalDevice] Getting information....")
	var deviceCount uint32
	result := vk.EnumeratePhysicalDevices(instance, &deviceCount, nil)
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Error getting physical device count: %v", result))
		return nil
	}
	fmt.Printf("\tFound total %v Physical Device(s).....", deviceCount)
	var physicalDevices = make([]vk.PhysicalDevice, deviceCount)
	result = vk.EnumeratePhysicalDevices(instance, &deviceCount, physicalDevices)
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Error getting physical device count: %v", result))
		return nil
	}
	return physicalDevices
}
