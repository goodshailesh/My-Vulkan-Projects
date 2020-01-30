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
	var pLogicalDevice *vk.Device
	var commandPool *vk.CommandPool
	var commandBuffers []vk.CommandBuffer

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
	pLogicalDevice = createDevice(physicalDevices[0], physicalDeviceFeatures)
	getInstanceLayerProperties()
	getDeviceLayerProperties(physicalDevices[0])
	getInstanceExtensionProperties()
	getDeviceExtensionProperties(physicalDevices[0])
	deviceWaitTillComplete(pLogicalDevice)
	commandPool = createCommandPool(*pLogicalDevice)
	commandBuffers = allocateCommandBuffers(*pLogicalDevice, *commandPool, 1)

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
	fmt.Printf("%T, %v", pLogicalDevice, pLogicalDevice)
	fmt.Println(commandPool)
	fmt.Println(commandBuffers)

	//Cleaningup code
	vk.FreeCommandBuffers(*pLogicalDevice, *commandPool, 1, commandBuffers)
	vk.DestroyCommandPool(*pLogicalDevice, *commandPool, nil)
}

func recordCommandIntoCommandBuffer(commandBuffer vk.CommandBuffer) {
	var commandBufferBeginInfo = vk.CommandBufferBeginInfo{
		SType: vk.StructureTypeCommandBufferBeginInfo,
		Flags: 0x0,
	}
	result := vk.BeginCommandBuffer(commandBuffer, &commandBufferBeginInfo)
	if result != vk.Success {
		fmt.Printf("Failed to begin command buffer with error : %v", result)
	}
}

func allocateCommandBuffers(pLogicalDevice vk.Device, pCommandPool vk.CommandPool, count uint32) []vk.CommandBuffer {
	fmt.Println("Allocating command buffer...............")
	var commandBuffers []vk.CommandBuffer
	commandBuffers = make([]vk.CommandBuffer, count)
	var commandBufferAllocateInfo = vk.CommandBufferAllocateInfo{
		SType:              vk.StructureTypeCommandBufferAllocateInfo,
		CommandPool:        pCommandPool,
		Level:              vk.CommandBufferLevelPrimary,
		CommandBufferCount: count,
	}
	result := vk.AllocateCommandBuffers(pLogicalDevice, &commandBufferAllocateInfo, commandBuffers)
	if result != vk.Success {
		fmt.Printf("Failed to create command buffer with error : %v", result)
	}
	return commandBuffers
}

func createCommandPool(pLogicalDevice vk.Device) *vk.CommandPool {
	fmt.Println("Creating Command Pool........")
	// Create CommandPool first
	var commandPool vk.CommandPool
	var commandPollCreateInfo = vk.CommandPoolCreateInfo{
		SType:            vk.StructureTypeCommandPoolCreateInfo,
		Flags:            vk.CommandPoolCreateFlags(vk.CommandPoolCreateResetCommandBufferBit),
		QueueFamilyIndex: 0,
	}
	result := vk.CreateCommandPool(pLogicalDevice, &commandPollCreateInfo, nil, &commandPool)
	if result != vk.Success {
		fmt.Printf("Failed to create command pool with error : %v", result)
	}
	return &commandPool
}

func createBuffer(device vk.Device, physicalDeviceProperties vk.PhysicalDeviceProperties) *vk.Buffer{
	var buffer vk.Buffer
	var bufferCreateInfo = vk.BufferCreateInfo{
		SType: vk.StructureTypeBufferCreateInfo,
		Flags: 0x0,
		Size: 1024, // 1Kb
		Usage: vk.BufferUsageTransferSrcBit | BufferUsageTransferDstBit, 
		SharingMode: SharingModeExclusive,		
	}
	result := vk.CreateBuffer(device, &bufferCreateInfo, nil, &buffer)
	if result != vk.Success {
		fmt.Printf("Failed to create buffer : %v", result)
	}
	return &buffer
}

func deviceWaitTillComplete(pLogicalDevice *vk.Device) {
	// Wait till the device is finished executing any work on behalf of your application
	// Wait on the host for the completion of outstanding queue operations for all queues on a given logical device
	fmt.Println("Waiting for the device to complete all pending queue operations.....")
	result := vk.DeviceWaitIdle(*pLogicalDevice)
	if result != vk.Success {
		fmt.Printf("Failed to wait for device to complete the execution with error : %v", result)
	}
	fmt.Println("All the queue operations are complete, device is free now....")
}
func getDeviceExtensionProperties(physicalDevice vk.PhysicalDevice) {
	fmt.Println("Listing available Extensions for device only..............")
	var propertyCount uint32
	vk.EnumerateDeviceExtensionProperties(physicalDevice, "", &propertyCount, nil)
	var pProperties []vk.ExtensionProperties = make([]vk.ExtensionProperties, propertyCount)
	vk.EnumerateDeviceExtensionProperties(physicalDevice, "", &propertyCount, pProperties)
	for i := 0; i < int(propertyCount); i++ {
		p := pProperties[i]
		p.Deref()
		fmt.Printf("\t*\tExtensionName: %v\n\t\tSpecVersion: %v\n", vk.ToString(p.ExtensionName[:]), p.SpecVersion)
	}
}

func getInstanceExtensionProperties() {
	fmt.Println("Listing available Extensions for instance only..............")
	var propertyCount uint32
	//vk.EnumerateInstanceLayerProperties(&propertyCount, nil)
	vk.EnumerateInstanceExtensionProperties("", &propertyCount, nil)
	var pProperties []vk.ExtensionProperties = make([]vk.ExtensionProperties, propertyCount)
	//vk.EnumerateInstanceLayerProperties(&propertyCount, pProperties)
	vk.EnumerateInstanceExtensionProperties("", &propertyCount, pProperties)
	for i := 0; i < int(propertyCount); i++ {
		p := pProperties[i]
		p.Deref()
		fmt.Printf("\t*\tExtensionName: %v\n\t\tSpecVersion: %v\n", vk.ToString(p.ExtensionName[:]), p.SpecVersion)
	}
}

func getDeviceLayerProperties(physicalDevice vk.PhysicalDevice) {
	fmt.Println("Listing available layers for device only..............")
	var propertyCount uint32
	vk.EnumerateDeviceLayerProperties(physicalDevice, &propertyCount, nil)
	var pProperties []vk.LayerProperties = make([]vk.LayerProperties, propertyCount)
	vk.EnumerateDeviceLayerProperties(physicalDevice, &propertyCount, pProperties)
	for i := 0; i < int(propertyCount); i++ {
		p := pProperties[i]
		p.Deref()
		fmt.Printf("\t*\tName: %v\n\t\tDescription: %v\n\t\tSpecVersion: %v\n\t\tImplementationVersion: %v\n", vk.ToString(p.LayerName[:]), vk.ToString(p.Description[:]), p.SpecVersion, p.ImplementationVersion)
	}
}

func getInstanceLayerProperties() {
	fmt.Println("Listing available layers for instance only..............")
	var propertyCount uint32
	vk.EnumerateInstanceLayerProperties(&propertyCount, nil)
	var pProperties []vk.LayerProperties = make([]vk.LayerProperties, propertyCount)
	vk.EnumerateInstanceLayerProperties(&propertyCount, pProperties)
	for i := 0; i < int(propertyCount); i++ {
		p := pProperties[i]
		p.Deref()
		fmt.Printf("\t*\tName: %v\n\t\tDescription: %v\n\t\tSpecVersion: %v\n\t\tImplementationVersion: %v\n", vk.ToString(p.LayerName[:]), vk.ToString(p.Description[:]), p.SpecVersion, p.ImplementationVersion)
	}
}

func createDevice(physicalDevice vk.PhysicalDevice, physicalDeviceFeatures vk.PhysicalDeviceFeatures) *vk.Device {
	fmt.Println("Creating Logical device..............")
	var logicalDevice vk.Device
	// var physicalDeviceFeatures []vk.PhysicalDeviceFeatures = []vk.PhysicalDeviceFeatures{
	// 	physicalDeviceFeatures,
	// }
	var pEnabledDeviceFeatures []vk.PhysicalDeviceFeatures = make([]vk.PhysicalDeviceFeatures, 1)
	deviceQueueCreateInfoSlice := []vk.DeviceQueueCreateInfo{{
		SType: vk.StructureTypeDeviceQueueCreateInfo,
		//QueueCount:       16,
		QueueCount: 1,
		//QueueFamilyIndex: 0,
		PQueuePriorities: []float32{1.0},
	}}
	//var deviceExtensions = []string{"VK_KHR_surface\x00"}
	//var deviceExtensions = []string{"VK_KHR_swapchain\x00"}
	//var deviceLayers = []string{"VK_LAYER_KHRONOS_validation\x00"}
	var deviceCreateInfo *vk.DeviceCreateInfo = &vk.DeviceCreateInfo{
		SType:                vk.StructureTypeDeviceCreateInfo,
		QueueCreateInfoCount: uint32(len(deviceQueueCreateInfoSlice)),
		PQueueCreateInfos:    deviceQueueCreateInfoSlice,
		// EnabledLayerCount:       uint32(len(deviceLayers)),
		// PpEnabledLayerNames:     deviceLayers,
		// EnabledExtensionCount:   uint32(len(deviceExtensions)),
		// PpEnabledExtensionNames: deviceExtensions,
		// Following will enable given features from slice
		//PEnabledFeatures:        physicalDeviceFeatures,
		// Following will return structure of enable features into the slice
		PEnabledFeatures: pEnabledDeviceFeatures,
	}
	result := vk.CreateDevice(physicalDevice, deviceCreateInfo, nil, &logicalDevice)
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Error creating Logical device: %v", result))
		return nil
	}
	fmt.Println("Created Logical device..............")
	return &logicalDevice
}

func uitlPrintDeviceQueueFamilyProperties(qfs []vk.QueueFamilyProperties) {
	fmt.Println("Print device queue properties..............")
	fmt.Printf("\tFound %v families\n", len(qfs))
	for idx, qf := range qfs {
		qf.Deref()
		fmt.Printf("\tFamily %v, Number of queues in the family (vk.QueueCount): %v\n", idx, qf.QueueCount)
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
	fmt.Printf("\tFound total %v Physical Device(s).....\n", deviceCount)
	var physicalDevices = make([]vk.PhysicalDevice, deviceCount)
	result = vk.EnumeratePhysicalDevices(instance, &deviceCount, physicalDevices)
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Error getting physical device count: %v", result))
		return nil
	}
	return physicalDevices
}
